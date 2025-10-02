{% for transaction in midaz_transaction.transaction %}
{{ transaction.id }} - TransactionID
{{ transaction.amount }} - Amount
{{ transaction.asset_code }} - Asset Code
{% endfor %}