<?xml version="1.0" encoding="UTF-8"?>
{%- if not transaction_id -%}
{% set transaction_id = "019649e7-0166-7d94-8a4c-c9016d8b2a16" %}
{%- endif -%}
{%- for t in transaction.transaction -%}
{%- if transaction_id == "" or t.id == transaction_id -%}
{% set total_movimentado = 0 %}
<Transacao>
    <Identificador>{{ t.id }}</Identificador>
    <Descricao>{{ t.description }}</Descricao>
    <Template>{{ t.template }}</Template>
    <DataCriacao>{{ t.created_at }}</DataCriacao>
    <Status>{{ t.status }}</Status>
    <Valor scale="{{ t.amount_scale|xmlattr }}">
        {{ t.amount }}
    </Valor>
    <Moeda>{{ t.asset_code }}</Moeda>
    <PlanoContas>{{ t.chart_of_accounts_group_name }}</PlanoContas>

    {% for org in onboarding.organization %}
    <Organizacao>
        <CNPJ>{{ org.legal_document }}</CNPJ>
        <NomeLegal>{{ org.legal_name }}</NomeLegal>
        <NomeFantasia>{{ org.doing_business_as }}</NomeFantasia>
        <Endereco>{{ org.address.line1 }}, {{ org.address.city }} - {{ org.address.state }}</Endereco>
    </Organizacao>
    {% endfor %}

    {% for l in onboarding.ledger %}
    <Ledger>
        <Nome>{{ l.name }}</Nome>
        <Status>{{ l.status }}</Status>
    </Ledger>
    {% endfor %}

    {% for a in onboarding.asset %}
    <Ativo>
        <Nome>{{ a.name }}</Nome>
        <Tipo>{{ a.type }}</Tipo>
        <Codigo>{{ a.code }}</Codigo>
    </Ativo>
    {% endfor %}

    <Operacoes>
        {% for operation in transaction.operation -%}
        {% if operation.transaction_id == transaction_id %}
            {% set total_movimentado = total_movimentado|add:operation.amount %}
            <Operacao>
                <ID>{{ operation.id }}</ID>
                <Descricao>{{ operation.description }}</Descricao>
                <Tipo>{{ operation.type }}</Tipo>
                <Conta>
                    <Alias>{{ operation.account_alias }}</Alias>
                </Conta>
                <Valor scale="{{ operation.amount_scale|xmlattr }}">{{ operation.amount }}</Valor>
                <SaldoDisponivelApos scale="{{ operation.balance_scale|xmlattr }}">
                    {{ operation.available_balance_after }}
                </SaldoDisponivelApos>
            </Operacao>
        {% endif %}
        {%- endfor %}
    </Operacoes>

    <TotalMovimentado>
        {{ total_movimentado }}
    </TotalMovimentado>
</Transacao>
{% endif %}
{%- endfor %}