<?xml version="1.0" encoding="UTF-8"?>
{% for t in midaz_transaction.transaction -%}
<Transacao>
    <Identificador>{{ t.id }}</Identificador>
    <Descricao>{{ t.description }}</Descricao>
    <DataCriacao>{{ t.created_at }}</DataCriacao>
    <Status>{{ t.status }}</Status>
    <Valor>
        {{ t.amount }}
    </Valor>
    <Moeda>{{ t.asset_code }}</Moeda>
    <PlanoContas>{{ t.chart_of_accounts_group_name }}</PlanoContas>
    <Mensagem>
        {{ midaz_transaction_metadata.transaction.0.metadata.mensagem }}
    </Mensagem>

    <Organizacao>
        <CNPJ>{{ midaz_onboarding.organization.0.legal_document }}</CNPJ>
        <NomeLegal>{{ midaz_onboarding.organization.0.legal_name }}</NomeLegal>
        <NomeFantasia>{{ midaz_onboarding.organization.0.doing_business_as }}</NomeFantasia>
        <Endereco>{{ midaz_onboarding.organization.0.address.line1 }}, {{ midaz_onboarding.organization.0.address.city }} - {{ midaz_onboarding.organization.0.address.state }}</Endereco>
    </Organizacao>

    {% for l in midaz_onboarding.ledger %}
    <Ledger>
        <Nome>{{ l.name }}</Nome>
        <Status>{{ l.status }}</Status>
    </Ledger>
    {% endfor %}

    {% for a in midaz_onboarding.asset %}
    <Ativo>
        <Nome>{{ a.name }}</Nome>
        <Tipo>{{ a.type }}</Tipo>
        <Codigo>{{ a.code }}</Codigo>
    </Ativo>
    {% endfor %}

    <Operacoes>
        {% for operation in midaz_transaction.operation -%}
        {% if operation.account_alias != "@external/BRL" && operation.transaction_id == t.id -%}
            <Operacao>
                <ID>{{ operation.id }}</ID>
                <Descricao>{{ operation.description }}</Descricao>
                <Tipo>{{ operation.type }}</Tipo>
                <Conta>
                    <Alias>{{ operation.account_alias }}</Alias>
                </Conta>
                <Valor>{{ operation.amount }}</Valor>
                <SaldoDisponivelApos>
                    {{ operation.available_balance_after }}
                </SaldoDisponivelApos>
                <Porcentagem>
                    {{ operation.amount|percent_of:t.amount }}
                </Porcentagem>
            </Operacao>
        {% endif %}
        {%- endfor %}
    </Operacoes>

</Transacao>
{% endfor %}

<TotalMovimentado>
    {% sum_by midaz_transaction.operation by "amount" if account_alias != "@external/BRL" %}
</TotalMovimentado>

<Totais>
    <Soma>
        {% sum_by midaz_transaction.operation by "amount" if account_alias != "@external/BRL" %}
    </Soma>
    <Contagem>
        {% count_by midaz_transaction.operation if account_alias != "@external/BRL" %}
    </Contagem>
    <Media>
        {% avg_by midaz_transaction.operation by "amount" if account_alias != "@external/BRL" %}
    </Media>
    <Minimo>
        {% min_by midaz_transaction.operation by "amount" if account_alias != "@external/BRL" %}
    </Minimo>
    <Maximo>
        {% max_by midaz_transaction.operation by "amount" if account_alias != "@external/BRL" %}
    </Maximo>
</Totais>