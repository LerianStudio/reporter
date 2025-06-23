# GET /v1/reports - Lista Todos os Relatórios

## 📋 Resumo da Implementação

Este documento descreve a implementação do endpoint `GET /v1/reports` seguindo todos os padrões Lerian e boas práticas do TRD.

## 🚀 Endpoint Implementado

```
GET /v1/reports
```

### Funcionalidades

- ✅ Listagem paginada de todos os relatórios
- ✅ Filtros por status (processing, finished, error)
- ✅ Filtros por templateId
- ✅ Filtros por data de criação (YYYY-MM-DD)
- ✅ Paginação com limit e page
- ✅ Ordenação por data de criação (mais recentes primeiro)
- ✅ Autenticação e autorização
- ✅ Isolamento por organização

### Parâmetros Query Suportados

| Parâmetro | Tipo | Descrição | Exemplo |
|-----------|------|-----------|---------|
| `status` | string | Status do relatório | `finished`, `processing`, `error` |
| `templateId` | UUID | ID do template | `019672b1-9d50-7360-9df5-5099dd166709` |
| `createdAt` | date | Data de criação | `2024-01-15` |
| `page` | int | Número da página | `1` (default) |
| `limit` | int | Itens por página | `10` (default, max 100) |

## 🏗️ Arquivos Implementados/Modificados

### 1. **Repository Layer**
- `pkg/mongodb/report/report.mongodb.go`
  - Adicionado método `FindList()` na interface `Repository`
  - Implementado `FindList()` com filtros e paginação
  - Suporte a filtros por status, templateId e data

### 2. **Service Layer**
- `components/manager/internal/services/get-all-reports.go`
  - Implementado `GetAllReports()` seguindo padrão dos templates
  - Tratamento de erros e logs
  - Telemetria/tracing

### 3. **HTTP Handler Layer**
- `components/manager/internal/adapters/http/in/report.go`
  - Adicionado `GetAllReports()` handler
  - Documentação Swagger completa
  - Validação de parâmetros
  - Paginação response

### 4. **Routes Configuration**
- `components/manager/internal/adapters/http/in/routes.go`
  - Adicionada rota `GET /v1/reports`
  - Middleware de autenticação/autorização

### 5. **Query Parameters**
- `pkg/net/http/http_utils.go`
  - Adicionados campos `Status` e `TemplateID` no `QueryHeader`
  - Validação de parâmetros templateId

### 6. **Tests**
- `components/manager/internal/services/get-all-reports_test.go`
  - Testes unitários completos
  - Coverage de cenários: sucesso, erro, filtros, paginação
  - Mocks gerados automaticamente

### 7. **Postman Collection**
- `components/manager/postman/Plugins Smart Templates.postman_collection.json`
  - Adicionada requisição "Get all reports" com exemplos

### 8. **Test Scripts**
- `scripts/test-get-all-reports.sh`
  - Script de teste manual do endpoint
  - Exemplos de todos os filtros disponíveis

## 🔧 Padrões Lerian Implementados

### ✅ **Observabilidade**
- **Telemetry/Tracing**: OpenTelemetry spans em todas as camadas
- **Logging**: Logs estruturados com contexto
- **Error Handling**: Tratamento consistente de erros

### ✅ **Security**
- **Authorization**: Middleware `auth.Authorize()` 
- **Organization Isolation**: Filtro automático por `organization_id`
- **Input Validation**: Validação de todos os parâmetros

### ✅ **Performance**
- **Pagination**: Limite máximo configurável
- **Database Optimization**: Indexes adequados para queries
- **Sorting**: Ordenação eficiente por data

### ✅ **Clean Architecture**
- **Separation of Concerns**: Repository → Service → Handler
- **Dependency Injection**: Interfaces bem definidas
- **Testing**: Mocks e testes unitários

### ✅ **API Standards**
- **REST Compliance**: GET para listagem com query params
- **HTTP Status Codes**: 200 para sucesso, 4xx/5xx para erros
- **Response Format**: Paginação padronizada
- **Documentation**: Swagger/OpenAPI completo

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

- **200 OK**: Lista retornada com sucesso (pode ser vazia)
- **400 Bad Request**: Parâmetros inválidos
- **401 Unauthorized**: Token ausente/inválido
- **403 Forbidden**: Sem permissão para o recurso
- **500 Internal Server Error**: Erro interno do servidor

## 🧪 Testes

### Unit Tests
```bash
go test ./components/manager/internal/services/ -v -run Test_getAllReports
```

### Integration Tests
```bash
./scripts/test-get-all-reports.sh
```

### Swagger Documentation
Acesse: `http://localhost:4005/swagger/index.html`

## 🔄 Exemplos de Uso

### 1. Listar todos os relatórios
```bash
curl -X GET "http://localhost:4005/v1/reports" \
  -H "X-Organization-Id: 01962525-a636-7a03-a2f2-5ef630c1f07e"
```

### 2. Filtrar por status
```bash
curl -X GET "http://localhost:4005/v1/reports?status=finished" \
  -H "X-Organization-Id: 01962525-a636-7a03-a2f2-5ef630c1f07e"
```

### 3. Paginação
```bash
curl -X GET "http://localhost:4005/v1/reports?page=2&limit=5" \
  -H "X-Organization-Id: 01962525-a636-7a03-a2f2-5ef630c1f07e"
```

### 4. Filtros combinados
```bash
curl -X GET "http://localhost:4005/v1/reports?status=finished&templateId=019672b1-9d50-7360-9df5-5099dd166709&page=1&limit=10" \
  -H "X-Organization-Id: 01962525-a636-7a03-a2f2-5ef630c1f07e"
```

## ✅ Compliance TRD

Esta implementação atende todos os requisitos do TRD 2-create-trd.mdc:

- ✅ **Dependências**: Usa lib-commons para logging/telemetry
- ✅ **Boas Práticas**: Clean Architecture, error handling, testing
- ✅ **Performance**: Paginação, indexes, ordenação eficiente
- ✅ **Security**: Autenticação, autorização, isolamento organizacional
- ✅ **Observabilidade**: Tracing, logging, metrics
- ✅ **Testabilidade**: Unit tests, mocks, integration tests
- ✅ **Documentação**: Swagger, exemplos, scripts de teste

---

**✨ Implementação completa e pronta para produção seguindo todos os padrões Lerian!** 