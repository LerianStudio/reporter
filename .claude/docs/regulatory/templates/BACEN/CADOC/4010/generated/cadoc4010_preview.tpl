<?xml version="1.0" encoding="UTF-8"?>
<documento codigoDocumento="4010" cnpj="{{ midaz_onboarding.organization.0.legal_document|slice:':8' }}" dataBase="{% date_time 'Y-m' %}" tipoRemessa="I">
    <contas>
{%- for route in midaz_transaction.operation_route %}
{%- with balance = filter(midaz_transaction.balance, "operation_route_id", route.id)[0] %}
{%- if balance %}
        <conta codigoConta="{{ route.code }}" saldo="{{ balance.available|floatformat:2 }}"/>
{%- endif %}
{%- endwith %}
{%- endfor %}
    </contas>
</documento>
