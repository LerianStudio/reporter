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

## Generate report RabbitMQ message

- Exchange: `template-engine.generate-report.exchange`
- Queue: `template-engine.generate-report.queue`
- Key: `template-engine.generate-report.key`

```json
{
   "id": "019538ee-deee-769c-8859-cbe84fce9af7",
   "type": "html",
   "fileUrl": "s3://client-reports-bucket/templates/report_ativos_21022025.txt",
   "mappedFields":[
      {
         "midaz":{
            "organization":["legal_name"],
            "ledger":["name","description"]
         }
      }
   ]
}
```

## Swagger Documentation

The boilerplate includes Swagger documentation that helps in visualizing and interacting with the API endpoints. You can access the documentation by running the project and navigating to `http://localhost:4000/swagger/index.html`.