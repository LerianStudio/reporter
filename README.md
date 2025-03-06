# K8s Addons Boilerplate

## Overview

This repository is a boilerplate for creating Go-based projects with Kubernetes addons. It provides a structure to help you start quickly with Go, Kubernetes, and microservices development. The boilerplate includes basic CRUD endpoints and Swagger documentation.
## Quick Start

1. **Clone the Repository:**
    ```bash
    git clone https://github.com/LerianStudio/k8s-golang-addons-boilerplate.git
    ```

2. **Install Dependencies:**
    ```bash
    cd k8s-addons-boilerplate
    go mod tidy
    ```

3. **Run the Server:**
    ```bash
   make up
    ```
   
4. **Access the API:**
   Visit `http://localhost:4000` to interact with the API.
   
## Endpoints

### `GET /v1/example`
- **Description:** Retrieve all Example records, optionally filtering by metadata.
- **Query Parameters:**
    - `limit` (integer, default: 10): Limit the number of results.
    - `page` (integer, default: 1): Specify the page number.
    - `start_date` (string): Filter records from this start date.
    - `end_date` (string): Filter records until this end date.
    - `sort_order` (string, enum: `asc`, `desc`): Order the results.
- **Response:** JSON containing paginated list of Example records.

### `POST /v1/example`
- **Description:** Create a new Example record.
- **Request Body:**
    - `name` (string, required): Name of the example (minimum length: 1).
    - `age` (integer): Age of the example.
- **Response:** JSON containing the created Example record.

### `GET /v1/example/{id}`
- **Description:** Retrieve an Example record by its ID.
- **Path Parameter:**
    - `id` (string, required): The ID of the Example record.
- **Response:** JSON containing the Example record.

### `DELETE /v1/example/{id}`
- **Description:** Delete an Example record by its ID.
- **Path Parameter:**
    - `id` (string, required): The ID of the Example record.
- **Response:** No content.

### `PATCH /v1/example/{id}`
- **Description:** Update an existing Example record.
- **Path Parameter:**
    - `id` (string, required): The ID of the Example record.
- **Request Body:**
    - `name` (string, optional): New name for the Example.
    - `age` (integer, optional): New age for the Example.
- **Response:** JSON containing the updated Example record.

## Swagger Documentation

The boilerplate includes Swagger documentation that helps in visualizing and interacting with the API endpoints. You can access the documentation by running the project and navigating to `http://localhost:4000/swagger/index.html`.