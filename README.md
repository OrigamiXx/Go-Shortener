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
- Docker support for containerized deployment

## Prerequisites

- Go 1.21 or later
- AWS CLI configured with appropriate credentials
- AWS SAM CLI (for local development)
- Docker (for local DynamoDB and containerized deployment)
- Docker Compose (for local development)

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
├── scripts/          # Deployment and utility scripts
├── Dockerfile        # Docker build instructions
├── docker-compose.yml # Docker Compose configuration
└── .dockerignore     # Docker ignore file
```

## Docker Deployment

### Building the Docker Image

1. Build the Docker image:
   ```bash
   docker build -t go-shortener .
   ```

2. Run the container:
   ```bash
   docker run -p 8080:8080 \
     -e AWS_REGION=us-east-1 \
     -e DYNAMODB_TABLE=url-counter \
     -e AWS_ACCESS_KEY_ID=your_access_key \
     -e AWS_SECRET_ACCESS_KEY=your_secret_key \
     go-shortener
   ```

### Using Docker Compose

1. Start the application:
   ```bash
   docker-compose up
   ```

2. Stop the application:
   ```bash
   docker-compose down
   ```

### EC2 Deployment

1. Install Docker on EC2:
   ```bash
   sudo yum update -y
   sudo yum install -y docker
   sudo service docker start
   sudo usermod -a -G docker ec2-user
   ```

2. Install Docker Compose:
   ```bash
   sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
   sudo chmod +x /usr/local/bin/docker-compose
   ```

3. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/Go-Shortener.git
   cd Go-Shortener
   ```

4. Build and run:
   ```bash
   docker-compose up -d
   ```

## Setup

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Create DynamoDB tables:
   ```bash
   # Create URL shortener table
   aws dynamodb create-table \
     --table-name url-shortener \
     --attribute-definitions AttributeName=ShortCode,AttributeType=S \
     --key-schema AttributeName=ShortCode,KeyType=HASH \
     --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5

   # Create counter table for generating short codes
   aws dynamodb create-table \
     --table-name url-counter \
     --attribute-definitions AttributeName=CounterName,AttributeType=S \
     --key-schema AttributeName=CounterName,KeyType=HASH \
     --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5

   # Initialize the counter
   aws dynamodb put-item \
     --table-name url-counter \
     --item '{"CounterName": {"S": "url_counter"}, "CurrentValue": {"N": "0"}}'
   ```

3. Deploy Lambda functions:
   ```bash
   sam build
   sam deploy --guided
   ```

## Build Process

1. Build the Lambda functions:
   ```bash
   # Build for Linux (required for AWS Lambda)
   GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/lambda/create/main.go
   GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/lambda/redirect/main.go
   
   # Or use SAM build (recommended)
   sam build
   ```

2. Package and deploy:
   ```bash
   sam package --output-template-file packaged.yaml
   sam deploy --template-file packaged.yaml --stack-name url-shortener --capabilities CAPABILITY_IAM
   ```

## Testing

### Local Testing

1. Start local DynamoDB:
   ```bash
   docker run -p 8000:8000 amazon/dynamodb-local
   ```

2. Create local DynamoDB table:
   ```bash
   aws dynamodb create-table \
     --endpoint-url http://localhost:8000 \
     --table-name url-shortener \
     --attribute-definitions AttributeName=ShortCode,AttributeType=S \
     --key-schema AttributeName=ShortCode,KeyType=HASH \
     --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
   ```

3. Start local API:
   ```bash
   sam local start-api --docker-network host
   ```

4. Test the API endpoints:
   ```bash
   # Create a short URL
   curl -X POST http://localhost:3000/create \
     -H "Content-Type: application/json" \
     -d '{"url": "https://example.com"}'

   # Use the short URL (will redirect to original URL)
   curl -L http://localhost:3000/{shortCode}
   ```

### Production Testing

1. Create a short URL:
   ```bash
   curl -X POST https://your-api-endpoint/prod/create \
     -H "Content-Type: application/json" \
     -d '{"url": "https://example.com"}'
   ```

2. Test the redirect:
   ```bash
   # Using curl
   curl -L https://your-api-endpoint/prod/{shortCode}

   # Or simply open in your browser
   https://your-api-endpoint/prod/{shortCode}
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

## TODO

- [ ] Improve short URL generation:
  - Replace UUID-based generation with a more compact algorithm
  - Consider using base62 encoding for shorter URLs
  - Implement custom length configuration
- [ ] Implement URL expiration:
  - Add TTL support in DynamoDB
  - Allow setting custom expiration times
- [ ] Add rate limiting:
  - Implement API Gateway throttling
  - Add per-user rate limits

## License

MIT License 