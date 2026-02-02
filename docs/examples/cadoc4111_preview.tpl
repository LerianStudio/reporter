<?xml version="1.0" encoding="UTF-8"?>
<documento codigoDocumento="4111" cnpj="{{ midaz_onboarding.organization.0.legal_document|slice:":8" }}" dataBase="{% date_time "YYYY-MM" %}" tipoRemessa="I">
  <contas>
{%- for op_route in midaz_transaction.operation_route %}
{%- if op_route.code %}
    <conta codigoConta="{{ op_route.code }}" saldoDia="{% sum_by midaz_transaction.operation by "available_balance_after" if route == op_route.id %}" />
{%- endif %}
{%- endfor %}
  </contas>
</documento>
