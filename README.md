# ads_api

## Description
This project is an API for managing advertisements, built with Go. It provides endpoints for creating, retrieving, updating, and posting ads, as well as managing users.

## Features
- CRUD operations for ads and users
- SQLite database for data storage
- Logging functionality
- Environment variable support

## Setup
1. Clone the repository
2. Install dependencies: `go mod download`
3. Create a `.env` file in the `cmd/app` directory with the following content:
   ```
   TELEGRAM_BOT_TOKEN=your_telegram_bot_token
   TELEGRAM_CHANNEL_ID=your_telegram_channel_id
   ```
4. Run the application: `go run cmd/app/main.go`

## API Endpoints
- POST /ads - Create a new ad
- GET /ads - Retrieve all ads
- GET /ads/{id} - Retrieve a specific ad
- PUT /ads/{id} - Update an ad
- POST /ads/{id}/post - Post an ad
- POST /users - Create a new user
- GET /users - Retrieve all users
- GET /users/{userid} - Retrieve a specific user
- PUT /users/{userid} - Update a user

## Technologies Used
- Go
- Gorilla Mux for routing
- SQLite for database
- godotenv for environment variable management

## Contributing
Contributions are welcome. Please open an issue or submit a pull request.

## License
This project is licensed under the MIT License.