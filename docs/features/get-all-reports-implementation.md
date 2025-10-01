# GET /v1/reports - List All Reports

## 📋 Implementation Summary

This document describes the implementation of the `GET /v1/reports` endpoint following all Lerian standards and TRD best practices.

## 🚀 Implemented Endpoint

```
GET /v1/reports
```

### Features

- ✅ Paginated listing of all reports
- ✅ Filters by status (processing, finished, error)
- ✅ Filters by templateId
- ✅ Filters by creation date (YYYY-MM-DD)
- ✅ Pagination with limit and page
- ✅ Sorting by creation date (most recent first)
- ✅ Authentication and authorization
- ✅ Organization isolation

### Supported Query Parameters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `status` | string | Report status | `finished`, `processing`, `error` |
| `templateId` | UUID | Template ID | `019672b1-9d50-7360-9df5-5099dd166709` |
| `createdAt` | date | Creation date | `2024-01-15` |
| `page` | int | Page number | `1` (default) |
| `limit` | int | Items per page | `10` (default, max 100) |

## 🏗️ Implemented/Modified Files

### 1. **Repository Layer**
- `pkg/mongodb/report/report.mongodb.go`
  - Added `FindList()` method to `Repository` interface
  - Implemented `FindList()` with filters and pagination
  - Support for status, templateId and date filters

### 2. **Service Layer**
- `components/manager/internal/services/get-all-reports.go`
  - Implemented `GetAllReports()` following template pattern
  - Error handling and logging
  - Telemetry/tracing

### 3. **HTTP Handler Layer**
- `components/manager/internal/adapters/http/in/report.go`
  - Added `GetAllReports()` handler
  - Complete Swagger documentation
  - Parameter validation
  - Pagination response

### 4. **Routes Configuration**
- `components/manager/internal/adapters/http/in/routes.go`
  - Added `GET /v1/reports` route
  - Authentication/authorization middleware

### 5. **Query Parameters**
- `pkg/net/http/http_utils.go`
  - Added `Status` and `TemplateID` fields to `QueryHeader`
  - templateId parameter validation

### 6. **Tests**
- `components/manager/internal/services/get-all-reports_test.go`
  - Complete unit tests
  - Coverage of scenarios: success, error, filters, pagination
  - Auto-generated mocks

### 7. **Postman Collection**
- `components/manager/postman/Plugins Smart Templates.postman_collection.json`
  - Added "Get all reports" request with examples

### 8. **Test Scripts**
- `scripts/test-get-all-reports.sh`
  - Manual endpoint testing script
  - Examples of all available filters

## 🔧 Implemented Lerian Standards

### ✅ **Observability**
- **Telemetry/Tracing**: OpenTelemetry spans across all layers
- **Logging**: Structured logging with context
- **Error Handling**: Consistent error treatment

### ✅ **Security**
- **Authorization**: `auth.Authorize()` middleware
- **Organization Isolation**: Automatic filtering by `organization_id`
- **Input Validation**: Validation of all parameters

### ✅ **Performance**
- **Pagination**: Configurable maximum limit
- **Database Optimization**: Adequate indexes for queries
- **Sorting**: Efficient sorting by date

### ✅ **Clean Architecture**
- **Separation of Concerns**: Repository → Service → Handler
- **Dependency Injection**: Well-defined interfaces
- **Testing**: Mocks and unit tests

### ✅ **API Standards**
- **REST Compliance**: GET for listing with query params
- **HTTP Status Codes**: 200 for success, 4xx/5xx for errors
- **Response Format**: Standardized pagination
- **Documentation**: Complete Swagger/OpenAPI

## 📊 Response Format

```json
{
  "items": [
    {
      "id": "019672b1-9d50-7360-9df5-5099dd166709",
      "templateId": "019672b1-9d50-7360-9df5-5099dd166710",
      "status": "finished",
      "ledgerId": ["019672b1-9d50-7360-9df5-5099dd166800"],
      "filters": {},
      "metadata": {},
      "completedAt": "2024-01-15T10:30:00Z",
      "createdAt": "2024-01-15T10:00:00Z",
      "updatedAt": "2024-01-15T10:30:00Z"
    }
  ],
  "page": 1,
  "limit": 10,
  "total": 25
}
```

## 🚦 Status Codes

- **200 OK**: List returned successfully (may be empty)
- **400 Bad Request**: Invalid parameters
- **401 Unauthorized**: Missing/invalid token
- **403 Forbidden**: No permission for resource
- **500 Internal Server Error**: Internal server error

## 🧪 Tests

### Unit Tests
```bash
go test ./components/manager/internal/services/ -v -run Test_getAllReports
```

### Integration Tests
```bash
./scripts/test-get-all-reports.sh
```

### Swagger Documentation
Access: `http://localhost:4005/swagger/index.html`

## 🔄 Usage Examples

### 1. List all reports
```bash
curl -X GET "http://localhost:4005/v1/reports" \
  -H "X-Organization-Id: 01962525-a636-7a03-a2f2-5ef630c1f07e"
```

### 2. Filter by status
```bash
curl -X GET "http://localhost:4005/v1/reports?status=finished" \
  -H "X-Organization-Id: 01962525-a636-7a03-a2f2-5ef630c1f07e"
```

### 3. Pagination
```bash
curl -X GET "http://localhost:4005/v1/reports?page=2&limit=5" \
  -H "X-Organization-Id: 01962525-a636-7a03-a2f2-5ef630c1f07e"
```

### 4. Combined filters
```bash
curl -X GET "http://localhost:4005/v1/reports?status=finished&templateId=019672b1-9d50-7360-9df5-5099dd166709&page=1&limit=10" \
  -H "X-Organization-Id: 01962525-a636-7a03-a2f2-5ef630c1f07e"
```

## ✅ TRD Compliance

This implementation meets all requirements of TRD 2-create-trd.mdc:

- ✅ **Dependencies**: Uses lib-commons for logging/telemetry
- ✅ **Best Practices**: Clean Architecture, error handling, testing
- ✅ **Performance**: Pagination, indexes, efficient sorting
- ✅ **Security**: Authentication, authorization, organizational isolation
- ✅ **Observability**: Tracing, logging, metrics
- ✅ **Testability**: Unit tests, mocks, integration tests
- ✅ **Documentation**: Swagger, examples, test scripts

---

**✨ Complete implementation ready for production following all Lerian standards!** 