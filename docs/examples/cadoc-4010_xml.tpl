<?xml version="1.0" encoding="UTF-8"?>
<documento codigoDocumento="4010" cnpj="{{ midaz_onboarding.organization.0.legal_document | slice:':8'}}"
dataBase="{% date_time "YYYY/MM" %}" tipoRemessa="I" >
<contas>
{%- for account in midaz_onboarding.account %}
{%- with balance = filter(midaz_transaction.balance, "account_id", account.id)[0] %}
<conta codigoConta="{{account.id}}" saldo="{{ balance.available}}" />
{%- endwith %}
{%- endfor%}

</contas>
</documento>
