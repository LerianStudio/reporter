{% if transaction_id is not defined %}
    {% set transaction_id = "" %}
{% endif %}

{% for trans in transaction %}
    {% if transaction_id == "" or trans.id == transaction_id %}
        {% set total_movimentado = 0 %}
        <Transacao>
            <Identificador>{{ trans.id }}</Identificador>
            <Descricao>{{ trans.description }}</Descricao>
            <Template>{{ trans.template }}</Template>
            <DataCriacao>{{ trans.created_at }}</DataCriacao>
            <Status>{{ trans.status }}</Status>
            <Valor scale="{{ trans.amount_scale }}">
                {{ trans.amount }}
            </Valor>
            <Moeda>{{ trans.asset_code }}</Moeda>
            <PlanoContas>{{ trans.chart_of_accounts_group_name }}</PlanoContas>

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
                {% for operation in trans.operation %}
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
    {% endif %}
{% endfor %}