# Go URL Shortener with AWS Lambda

A serverless URL shortener service built with Go and AWS Lambda. This service uses DynamoDB to store the URL mappings and provides two main endpoints:
- Create short URL
- Redirect to original URL

## Features

- Serverless architecture using AWS Lambda
- DynamoDB for data storage
- RESTful API endpoints
- Automatic URL validation
- Unique short code generation

## Prerequisites

- Go 1.21 or later
- AWS CLI configured with appropriate credentials
- AWS SAM CLI (for local development)

## Project Structure

```
.
├── cmd/
│   └── lambda/
│       ├── create/    # Create short URL Lambda function
│       └── redirect/  # Redirect Lambda function
├── internal/
│   ├── models/       # Data models
│   └── storage/      # DynamoDB storage implementation
├── pkg/
│   └── shortener/    # URL shortener logic
└── scripts/          # Deployment and utility scripts
```

## Setup

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Create DynamoDB table:
   ```bash
   aws dynamodb create-table \
     --table-name url-shortener \
     --attribute-definitions AttributeName=ShortCode,AttributeType=S \
     --key-schema AttributeName=ShortCode,KeyType=HASH \
     --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
   ```

3. Deploy Lambda functions:
   ```bash
   sam build
   sam deploy --guided
   ```

## API Endpoints

### Create Short URL
- Method: POST
- Path: /create
- Request Body:
  ```json
  {
    "url": "https://example.com"
  }
  ```
- Response:
  ```json
  {
    "shortCode": "abc123",
    "shortUrl": "https://your-domain.com/abc123"
  }
  ```

### Redirect
- Method: GET
- Path: /{shortCode}
- Response: 302 Redirect to original URL

## Local Development

1. Start local DynamoDB:
   ```bash
   docker run -p 8000:8000 amazon/dynamodb-local
   ```

2. Run tests:
   ```bash
   go test ./...
   ```

3. Test locally:
   ```bash
   sam local start-api
   ```

## License

MIT License 