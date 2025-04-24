<?xml version="1.0" encoding="UTF-8"?>
{% for t in transaction.transaction -%}
<Transacao>
    <Identificador>{{ t.id }}</Identificador>
    <Descricao>{{ t.description }}</Descricao>
    <Template>{{ t.template }}</Template>
    <DataCriacao>{{ t.created_at }}</DataCriacao>
    <Status>{{ t.status }}</Status>
    <Valor scale="{{ t.amount_scale }}">
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
        {% if operation.account_alias != "@external/BRL" %}
            <Operacao>
                <ID>{{ operation.id }}</ID>
                <Descricao>{{ operation.description }}</Descricao>
                <Tipo>{{ operation.type }}</Tipo>
                <Conta>
                    <Alias>{{ operation.account_alias }}</Alias>
                </Conta>
                <Valor scale="{{ operation.amount_scale }}">{{ operation.amount }}</Valor>
                <SaldoDisponivelApos scale="{{ operation.balance_scale_after }}">
                    {{ operation.available_balance_after|scale:operation.balance_scale_after }}
                </SaldoDisponivelApos>
                <Porcentagem>
                    {{ operation.amount|percent_of:t.amount }}
                </Porcentagem>
            </Operacao>
        {% endif %}
        {%- endfor %}
    </Operacoes>

    <TotalMovimentado>
        {% sum_by transaction.operation by "amount" if account_alias != "@external/BRL" scale 2 %}
    </TotalMovimentado>

    <Totais>
        <Soma>
            {% sum_by transaction.operation by "amount" if account_alias != "@external/BRL" scale 2 %}
        </Soma>
        <Contagem>
            {% count_by transaction.operation if account_alias != "@external/BRL" %}
        </Contagem>
        <Media>
            {% avg_by transaction.operation by "amount" if account_alias != "@external/BRL" scale 2 %}
        </Media>
        <Minimo>
            {% min_by transaction.operation by "amount" if account_alias != "@external/BRL" scale 2 %}
        </Minimo>
        <Maximo>
            {% max_by transaction.operation by "amount" if account_alias != "@external/BRL" scale 2 %}
        </Maximo>
    </Totais>

</Transacao>
{% endfor %}