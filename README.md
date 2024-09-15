# File Sharing Platform

This project is a file-sharing platform built using the Go language and the Gin framework. It allows users to upload files, search for files, and share. The project integrates with AWS S3 for file storage, PostgreSQL for metadata storage, and Redis for caching. 

## Features

1. User authentication using JWT tokens
2. File Upload to AWS S3: Users can upload files, and the metadata is stored in PostgreSQL.
3. Search Functionality: Files can be searched by name, upload date, or file type.
4. File Caching: Cached responses for file searches using Redis to improve performance.
5. Concurrency for processing large uploads using goroutines, channels and semaphore
6. Rate Limiting: Limits users to 100 requests per minute to protect against abuse.
7. A background worker that periodically deletes expired files from S3 and
 their metadata from the database.

## Running the project

1. Clone the repository
2. Setup .env file
3. If you're on Windows, you need to install and run Docker Desktop
4. Run `docker-compose up --build`
5. The application will be available at `http://localhost:8080`

## Environment Variables

Create a `.env` file in the root directory of the project with the following variables:

### .env

```env

PORT=8080
DB_HOST="db"
DB_USER="db_user"
DB_PASSWORD="db_pass"
DB_NAME="file_sharing"
REDIS_HOST="redis"
REDIS_ADDR="redis:6379"
JWT_SECRET="your_secret"

AWS_ACCESS_KEY_ID="your_access_key_id"
AWS_SECRET_ACCESS_KEY="your_secret_access_key"
AWS_REGION="Your_region"
S3_BUCKET="your_bucket_name"
```

## API Endpoints

### User Registration
URL: `/user/register`
Method: `POST`
Request Body:
```json
{
 "email": "akarsh1@gmail.com",
 "password": "pass1"
}
```

### User Login
URL: `/user/login`
Method: `POST`
Request Body:
```json
{
 "email": "akarsh1@gmail.com",
 "password": "pass1"
}
```

### File Upload
URL: `/file/upload`
Method: `POST`
Authorization: JWT token as bearer token
Request Body:
In form data: key:file value:(selected file)

### Get Uploaded files
URL: `/file/`
Method: `GET`
Authorization: JWT token as bearer token

### Get Uploaded files' metadata
URL: `/file/`
Method: `GET`
Authorization: JWT token as bearer token

### Get File URL from file id
URL: `/file/:id`
Method: `GET`
Authorization: JWT token as bearer token

### Search file by name/date/type
URL: `/file/search`
Method: `GET`
Authorization: JWT token as bearer token
Params: Query parameter with key as either name, date or type

## Concurrency for large uploads
To implement concurrency for processing large file uploads using goroutines, upload process is broken into multiple steps, such as reading the file in chunks and then processing or uploading those chunks concurrently using goroutines. This approach ensures that the system can handle large files without blocking the server for too long. The goroutines share data with the main routine using a channel and the main routine is put on hold during the execution of other goroutines using a semaphore.

## Rate limiting using Redis Cache
The number of requests per user(tracked using their IP address) is stored in the redis cache which is reset every minute. If the limit is exceeded, server responds with an HTTP status 429 Too Many Requests.

## Deployment on Render
URL: `https://two1bce2661-backend.onrender.com`
