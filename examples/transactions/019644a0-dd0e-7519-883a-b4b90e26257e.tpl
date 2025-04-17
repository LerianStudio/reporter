{% set total_movimentado = 0 %}
<Transacao>
    <Identificador>{{ transaction.transaction.id }}</Identificador>
    <Descricao>{{ transaction.transaction.description }}</Descricao>
    <Template>{{ transaction.transaction.template }}</Template>
    <DataCriacao>{{ transaction.transaction.created_at }}</DataCriacao>
    <Status>{{ transaction.transaction.status }}</Status>
    <Valor scale="{{ transaction.transaction.amount_scale }}">
        {{ transaction.transaction.amount }}
    </Valor>
    <Moeda>{{ transaction.transaction.asset_code }}</Moeda>
    <PlanoContas>{{ transaction.transaction.chart_of_accounts_group_name }}</PlanoContas>

    <Organizacao>
        <CNPJ>{{ onboarding.organization.legal_document }}</CNPJ>
        <NomeLegal>{{ onboarding.organization.legal_name }}</NomeLegal>
        <NomeFantasia>{{ onboarding.organization.doing_business_as }}</NomeFantasia>
        <Endereco>{{ onboarding.organization.address }}</Endereco>
    </Organizacao>

    <Ledger>
        <Nome>{{ onboarding.ledger.name }}</Nome>
        <Status>{{ onboarding.ledger.status }}</Status>
    </Ledger>

    <Ativo>
        <Nome>{{ onboarding.asset.name }}</Nome>
        <Tipo>{{ onboarding.asset.type }}</Tipo>
        <Codigo>{{ onboarding.asset.code }}</Codigo>
    </Ativo>

    <Operacoes>
        {% for operation in transaction.operation %}
            {% set total_movimentado = total_movimentado|add:operation.amount %}
            <Operacao>
                <ID>{{ operation.id }}</ID>
                <Descricao>{{ operation.description }}</Descricao>
                <Tipo>{{ operation.type }}</Tipo>
                <Conta>
                    <Nome>{{ Portfolio.account.name }}</Nome>
                    <Alias>{{ operation.account_alias }}</Alias>
                </Conta>
                <Valor scale="{{ operation.amount_scale }}">{{ operation.amount }}</Valor>
                <SaldoDisponivelApos scale="{{ operation.balance_scale }}">
                    {{ operation.available_balance_after }}
                </SaldoDisponivelApos>
            </Operacao>
        {% endfor %}
    </Operacoes>

    <TotalMovimentado>
        {{ total_movimentado }}
    </TotalMovimentado>
</Transacao>
