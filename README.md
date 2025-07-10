# Mowsy API

A comprehensive Go REST API for Mowsy, a hyperlocal gig economy platform for yard work jobs and equipment rentals.

## Features

- **User Management**: Registration, authentication, profile management
- **Job Management**: Create, browse, and manage yard work jobs
- **Equipment Rental**: List and rent lawn equipment
- **Payment Processing**: Stripe integration for secure payments
- **Location Services**: Geocoding with elementary school district filtering
- **File Upload**: S3 integration for images and documents
- **Admin Panel**: Administrative functions and statistics

## Tech Stack

- **Language**: Go 1.21
- **Framework**: Gin
- **Database**: PostgreSQL with GORM
- **Authentication**: JWT tokens
- **Payments**: Stripe
- **Storage**: AWS S3
- **Geocoding**: Geocodio API
- **Deployment**: AWS Lambda + API Gateway

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- AWS account (for S3 and Lambda deployment)
- Stripe account
- Geocodio API key

### Environment Variables

Create a `.env` file with the following variables:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mowsy_db
DB_USER=postgres
DB_PASSWORD=your_password

# JWT
JWT_SECRET=your_super_secret_jwt_key

# Stripe
STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key
STRIPE_PUBLISHABLE_KEY=pk_test_your_stripe_publishable_key

# Geocodio
GEOCODIO_API_KEY=your_geocodio_api_key

# AWS
AWS_REGION=us-east-1
AWS_S3_BUCKET_NAME=mowsy-uploads

# Admin
ADMIN_API_KEY=your_admin_api_key

# Server
PORT=8080
GIN_MODE=debug
```

### Installation

1. Clone the repository:
```bash
git clone https://github.com/your-org/mowsy-api.git
cd mowsy-api
```

2. Install dependencies:
```bash
go mod download
```

3. Set up your database and environment variables

4. Run the application locally:
```bash
go run cmd/local/main.go
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/logout` - Logout user

### User Management
- `GET /api/v1/users/me` - Get current user profile
- `PUT /api/v1/users/me` - Update user profile
- `POST /api/v1/users/me/insurance` - Upload insurance document
- `GET /api/v1/users/:id/reviews` - Get user reviews
- `GET /api/v1/users/:id/profile` - Get public user profile

### Jobs
- `GET /api/v1/jobs` - List jobs (with filters)
- `POST /api/v1/jobs` - Create new job
- `GET /api/v1/jobs/:id` - Get job details
- `PUT /api/v1/jobs/:id` - Update job
- `DELETE /api/v1/jobs/:id` - Delete job
- `POST /api/v1/jobs/:id/apply` - Apply for job
- `GET /api/v1/jobs/:id/applications` - Get job applications
- `PUT /api/v1/jobs/:id/applications/:app_id` - Update application status
- `POST /api/v1/jobs/:id/complete` - Mark job as completed

### Equipment
- `GET /api/v1/equipment` - List equipment (with filters)
- `POST /api/v1/equipment` - Add new equipment
- `GET /api/v1/equipment/:id` - Get equipment details
- `PUT /api/v1/equipment/:id` - Update equipment
- `DELETE /api/v1/equipment/:id` - Delete equipment
- `POST /api/v1/equipment/:id/rent` - Request equipment rental
- `GET /api/v1/equipment/:id/rentals` - Get rental requests
- `PUT /api/v1/equipment/:id/rentals/:rental_id` - Update rental status
- `POST /api/v1/equipment/rentals/:rental_id/complete` - Complete rental

### Payments
- `POST /api/v1/payments/create-intent` - Create payment intent
- `POST /api/v1/payments/confirm` - Confirm payment
- `GET /api/v1/payments/history` - Get payment history
- `GET /api/v1/payments/:id` - Get payment details

### File Upload
- `POST /api/v1/upload/image` - Upload image
- `POST /api/v1/upload/presigned-url` - Get presigned upload URL
- `DELETE /api/v1/upload/file` - Delete file

### Admin (requires X-Admin-Key header)
- `GET /api/v1/admin/stats` - Get platform statistics
- `GET /api/v1/admin/users` - List all users
- `PUT /api/v1/admin/users/:id/deactivate` - Deactivate user
- `PUT /api/v1/admin/users/:id/activate` - Activate user
- `PUT /api/v1/admin/users/:id/verify-insurance` - Verify insurance
- `DELETE /api/v1/admin/jobs/:id` - Remove job
- `DELETE /api/v1/admin/equipment/:id` - Remove equipment

## Database Schema

The API uses PostgreSQL with the following main tables:

- `users` - User accounts and profiles
- `jobs` - Job postings
- `job_applications` - Job applications
- `equipment` - Equipment listings
- `equipment_rentals` - Equipment rental requests
- `reviews` - User reviews
- `payments` - Payment records

## Location Features

The API supports hyperlocal filtering by:
- ZIP code
- Elementary school district (via Geocodio API)

## Security Features

- JWT-based authentication
- Password hashing with bcrypt
- Rate limiting
- Input validation and sanitization
- CORS configuration
- Insurance verification requirements

## Deployment

### AWS Lambda

1. Build the Lambda function:
```bash
GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/lambda/main.go
zip lambda-deployment.zip bootstrap
```

2. Deploy using AWS CLI or infrastructure as code tools

### Local Development

Run the local server:
```bash
go run cmd/local/main.go
```

## Testing

Run tests:
```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.