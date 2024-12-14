API Endpoints
1. Upload Employee Data (Excel)
Endpoint: POST /upload
Description: This endpoint allows you to upload an Excel file containing employee data. The file is processed asynchronously and stored in the database and Redis cache.
2. Get All Employee Data
Endpoint: GET /employees
Description: This endpoint fetches all employee data. The data is first checked in Redis. If not found, it will be fetched from the MySQL database and cached in Redis for 5 minutes.

3. Edit Employee Data
Endpoint: PUT /employee/:id
Description: This endpoint allows you to update the data for a specific employee using their ID.
