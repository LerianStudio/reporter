<RelatorioAnalitico>
    {%- with org = midaz_onboarding.organization[0] %}
    <Organizacao>{{ org.legal_name }} - CNPJ: {{ org.legal_document }}</Organizacao>
    {%- endwith %}
    <DataGeracao>28.04.2025</DataGeracao>
    {%- with ledger = midaz_onboarding.ledger[0] %}
    <Ledger>{{ ledger.name }}</Ledger>
    {%- endwith %}
    {%- for account in midaz_onboarding.account %}
    <Conta>
        <IDConta>{{ account.id }}</IDConta>
        <Alias>{{ account.alias }}</Alias>
        {%- with balance = filter(midaz_transaction.balance, "account_id", account.id)[0] %}
        <SaldoAtual>Usando filtro customizado: {{ balance.available|scale:balance.scale }}</SaldoAtual>
        {%- endwith %}
        {%- for balance in midaz_transaction.balance %}
        {%- if balance.account_id == account.id %}
        <SaldoAtual>Usando padrão pongo2: {{ balance.available|scale:balance.scale }}</SaldoAtual>
        {%- endif %}
        {%- endfor %}
        <Moeda>{{ account.asset_code }}</Moeda>
        <Operacoes>
        {%- for operation in midaz_transaction.operation %}
        {%- if operation.account_id == account.id %}
            {%- set valor_original = operation.amount|scale:operation.amount_scale %}
            {%- set valor_desconto = valor_original * 0.03 %}
            {%- set valor_final = valor_original - valor_desconto %}
            <Operacao>
                <IDOperacao>{{ operation.id }}</IDOperacao>
                <Descricao>{{ operation.description }}</Descricao>
                <Tipo>{{ operation.type }}</Type>
                <PlanoContas>{{ operation.chart_of_accounts }}</PlanoContas>
                <ValorOriginal>{{ valor_original }}</ValorOriginal>
                <ValorDesconto10porcento>{{ valor_desconto }}</ValorDesconto10porcento>
                <ValorFinalComDesconto>{{ valor_final }}</ValorFinalComDesconto>
                <Moeda>{{ operation.asset_code }}</Moeda>
                <Status>{{ operation.status }}</Status>
            </Operacao>
        {%- endif %}
        {%- endfor %}
        </Operacoes>
        <ResumoConta>
            <TotalOperacoes>{% count_by midaz_transaction.operation if account_id == account.id %}</TotalOperacoes>
            <SomatorioOperacoes>{% sum_by midaz_transaction.operation by "amount" if account_id == account.id scale 2 %}</SomatorioOperacoes>
            <MediaOperacoes>{% avg_by midaz_transaction.operation by "amount" if account_id == account.id scale 2 %}</MediaOperacoes>
        </ResumoConta>
    </Conta>
    {%- endfor %}
</RelatorioAnalitico>