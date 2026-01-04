import os
import logging
import argparse
from pymongo import MongoClient
from dotenv import load_dotenv
from datetime import datetime, UTC

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

# Load environment variables
load_dotenv()

MONGO_URI = os.environ.get("MONGO_URI")
MONGO_DB_NAME = os.environ.get("MONGO_DB_NAME", "articles_db")
MONGO_COLLECTION_NAME = os.environ.get("MONGO_COLLECTION_NAME", "articles")


def migrate_events(dry_run=False):
    """
    Migrates existing MongoDB events to the new standardized schema.
    """
    if not MONGO_URI:
        logger.error("MONGO_URI not found in environment variables.")
        return

    try:
        client = MongoClient(MONGO_URI)
        db = client[MONGO_DB_NAME]
        collection = db[MONGO_COLLECTION_NAME]

        logger.info(f"Connected to MongoDB: {MONGO_DB_NAME}.{MONGO_COLLECTION_NAME}")
        if dry_run:
            logger.info("Running in DRY RUN mode. No changes will be saved.")

        # Find all documents that don't have a 'payload' field (legacy format)
        query = {"payload": {"$exists": False}}
        total_docs = collection.count_documents(query)

        if total_docs == 0:
            logger.info("No legacy documents found. Migration not needed.")
            return

        logger.info(f"Found {total_docs} legacy documents to migrate.")

        migrated_count = 0
        error_count = 0
        example_count = 0

        cursor = collection.find(query)

        for doc in cursor:
            try:
                new_fields = {}
                unset_fields = {}

                # 1. Standardize Timestamp (extracted_at -> timestamp)
                if "extracted_at" in doc:
                    new_fields["timestamp"] = doc["extracted_at"]
                    unset_fields["extracted_at"] = ""
                else:
                    new_fields["timestamp"] = datetime.now(UTC).isoformat()

                # 2. Initialize Payload
                payload = {}

                # 3. Handle Extraction Events (article data)
                if "article" in doc:
                    article_data = doc["article"]
                    payload["title"] = article_data.get("title")
                    payload["link"] = article_data.get("link")
                    payload["published_date"] = article_data.get("published_date")
                    unset_fields["article"] = ""

                # 4. Handle Domain
                if "domain" in doc:
                    payload["domain"] = doc["domain"]
                    unset_fields["domain"] = ""

                # 5. Handle Error Events (error -> payload)
                if "error" in doc:
                    error_data = doc["error"]
                    if "message" in error_data:
                        payload["message"] = error_data["message"]
                    if "url" in error_data:
                        payload["url"] = error_data["url"]
                    if "traceback" in error_data:
                        payload["traceback"] = error_data["traceback"]

                    unset_fields["error"] = ""

                # 6. Handle Metadata (metadata -> meta)
                if "metadata" in doc:
                    new_fields["meta"] = doc["metadata"]
                    unset_fields["metadata"] = ""

                # 7. Assign Payload
                if payload:
                    new_fields["payload"] = payload

                # 8. Ensure Status
                if doc.get("status") != "ingested":
                    new_fields["status"] = "ingested"

                # Perform Update
                if new_fields:
                    update_op = {"$set": new_fields}
                    if unset_fields:
                        update_op["$unset"] = unset_fields

                    if dry_run:
                        if example_count < 5:
                            logger.info(
                                f"Dry Run - Would update document {doc['_id']}: {update_op}"
                            )
                            example_count += 1
                    else:
                        collection.update_one({"_id": doc["_id"]}, update_op)

                    migrated_count += 1

                    if not dry_run and migrated_count % 100 == 0:
                        logger.info(f"Migrated {migrated_count} documents...")

            except Exception as e:
                logger.error(f"Error migrating document {doc.get('_id')}: {e}")
                error_count += 1

        logger.info(
            f"Migration completed. Success: {migrated_count}, Errors: {error_count}"
        )

    except Exception as e:
        logger.error(f"Migration failed: {e}")
    finally:
        if client:
            client.close()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Migrate MongoDB events to standard schema."
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simulate migration without writing changes.",
    )
    args = parser.parse_args()

    if not args.dry_run:
        print(
            "WARNING: This script will modify data in your MongoDB article collection."
        )
        confirmation = input("Type 'yes' to proceed: ")
        if confirmation.lower() != "yes":
            logger.info("Migration cancelled.")
            exit(0)

    migrate_events(dry_run=args.dry_run)
