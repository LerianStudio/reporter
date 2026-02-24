<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Analytical Financial Report</title>
  <style>
    body {
      font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
      background-color: #f4f6f9;
      color: #333;
      padding: 40px;
    }

    .container {
      max-width: 900px;
      margin: auto;
      background: #ffffff;
      padding: 30px 40px;
      border-radius: 8px;
      box-shadow: 0 0 10px rgba(0,0,0,0.1);
    }

    h1 {
      text-align: center;
      font-size: 24px;
      color: #004080;
      margin-bottom: 10px;
    }

    .subtitle {
      text-align: center;
      font-size: 14px;
      color: #777;
      margin-bottom: 30px;
    }

    .section-title {
      font-weight: bold;
      font-size: 16px;
      margin-top: 30px;
      margin-bottom: 10px;
      color: #004080;
      border-bottom: 1px solid #ccc;
      padding-bottom: 4px;
    }

    table {
      width: 100%;
      border-collapse: collapse;
      margin-top: 15px;
    }

    th, td {
      padding: 10px;
      text-align: left;
      border: 1px solid #ccc;
    }

    th {
      background-color: #e8f0fe;
      font-weight: bold;
      color: #003366;
    }

    .info p {
      margin: 4px 0;
    }

    .footer {
      margin-top: 40px;
      text-align: center;
      font-size: 12px;
      color: #999;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>Analytical Financial Report</h1>
    <div class="subtitle">Detailed analysis of movements by linked accounts.</div>

    <div class="info">
      <p><strong>Generation Date:</strong> {% date_time "dd/MM/YYYY HH:mm" %}</p>
      <p><strong>Organization:</strong> {{ midaz_onboarding.organization.0.legal_name }}</p>
      <p><strong>Ledger:</strong> {{ midaz_onboarding.ledger.0.name }}</p>
    </div>

    {% for account in midaz_onboarding.account %}
      {% with balance = filter(midaz_transaction.balance, "account_id", account.id)[0] %}
        <div class="section-title">Account: {{ account.alias }}</div>
        <div class="info">
          <p><strong>ID:</strong> {{ account.id }}</p>
          <p><strong>Currency:</strong> {{ balance.asset_code }}</p>
          <p><strong>Current Balance:</strong> {{ balance.available }}</p>
        </div>

        <table>
          <thead>
            <tr>
              <th>Operation ID</th>
              <th>Type</th>
              <th>Original Amount</th>
              <th>Discount (3%)</th>
              <th>Final Amount</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            {% for operation in midaz_transaction.operation %}
              {% if operation.account_id == account.id %}
                {% set original_amount = operation.amount %}
                {% set discount_amount = original_amount * 0.03  %}
                {% set final_amount = original_amount - discount_amount %}
                <tr>
                  <td>{{ operation.id }}</td>
                  <td>{{ operation.type }}</td>
                  <td>{{ original_amount|floatformat:2 }}</td>
                  <td>{{ discount_amount|floatformat:2 }}</td>
                  <td>{{ final_amount|floatformat:2 }}</td>
                  <td>{{ operation.description }}</td>
                </tr>
              {% endif %}
            {% endfor %}
          </tbody>
        </table>
      {% endwith %}
    {% endfor %}

    <div class="footer">
      Document generated automatically via Smart Templates - Lerian · {{ midaz_onboarding.organization.0.legal_document }}
    </div>
  </div>
</body>
</html>