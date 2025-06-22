# ABOUT

A simple REST API used to check the exchange rate between currencies and to calculate the exchange value for selected currencies.  
The application fetches currency rates from the openexchangerates.org API.

---
# DEVELOPMENT

To run the application, the OpenExchange API requires an `app_id` linked to your user account.  
Create an account at https://openexchangerates.org/signup/free to receive your `app_id`.

Clone the repository:

```
git clone git@github.com:i-oles/currency-api.git
cd currency-api
```

Then using Docker (don't forget to pass ENV - your app_id):
```
docker build -t currency-api .
docker run -e APP_ID=<your_app_id> -p 8080:8080 currency-api
```

By default, the application uses the config located in `./config/dev.json`

```json
{
  "ListenAddress": ":8080",
  "ReadTimeout": 5,
  "WriteTimeout": 10,
  "ContextTimeout": 5,
  "APIURL": "https://openexchangerates.org/api/",
  "LogErrors": true
}
```

An example test request to the OpenExchange API is located in `./example`

The application also uses ***makefile***  
You can use following commands:

```

make build  --> for build aplication
make run    --> for running aplication locally
make test   --> for run tests
make lint   --> for run linter

```

# RUNNING

### GET /rates

Returns all possible exchange rate pairs between the requested currencies.  
This endpoint requires one parameter:

`currencies` - the currencies for which we want to get the exchange rates.

The result is returned rounded to 8 decimal places.  
In case of an error, the application returns an empty body and a status code 400.  
If the OpenExchangeRates API returns an error, the application also returns status code 400 and an empty body.

---
`GET /rates?currencies=GBP,USD`

```
--> Status: 200

[
    { "from": "USD", "to": "GBP", "rate": 0.74365300 },
    { "from": "GBP", "to": "USD", "rate": 1.34471319 },
]
```
---

`GET /rates?currencies=BTC,INR,LYD`

```
--> Status: 200

[
    {"from":"BTC","to":"INR","rate":8616935.06861531},
    {"from":"BTC","to":"LYD","rate":542699.05871178},
    {"from":"INR","to":"BTC","rate":0.00000012},
    {"from":"INR","to":"LYD","rate":0.06298052},
    {"from":"LYD","to":"BTC","rate":0.00000184},
    {"from":"LYD","to":"INR","rate":15.87792522}
]
```

---
Failure when only one currency is provided:

`GET /rates?currencies=BTC`

```
--> Status: 400
```
---
Failure when ***currencies*** parameter is empty:  

`GET /rates?currencies=`

```
--> Status: 400
```
---
Failure when the currency is not found in the OpenExchangeRates API:

`GET /rates?currencies=ABRAKADABRA`

```
--> Status: 400
```
---



### GET /exchange

Calculates the exchange value from one cryptocurrency to another.  
This endpoint requires three parameters:

- `from` - the cryptocurrency we want to exchange
- `to` - the cryptocurrency we want to receive
- `amount` - the amount of cryptocurrency to exchange

The data is returned based on the table below.  
The "Decimal places" column defines the precision to which the result is returned.  
In case of an error, the application returns an empty body and a status code 400.  

| CryptoCurrency | Decimal places | Rate (to USD) |
| ----------- | ----------- | ----------- |
| BEER | 18 | 0.00002461$
| FLOKI | 18 | 0.0001428$
| GATE| 18 | 6.87$
| USDT | 6 | 0.999$
| WBTC | 8 | 57,037.22$

---
`GET /exchange?from=WBTC&to=USDT&amount=1.0`

```
--> Status: 200

{"from":"WBTC","to":"USDT","amount":57094.314314}
```
---
`GET /exchange?from=GATE&to=WBTC&amount=12.0`

```
--> Status: 200

{"from":"GATE","to":"WBTC","amount":0.00144537}
```
---
`GET /exchange?from=FLOKI&to=BEER&amount=123.23`

```
--> Status: 200

{"from":"FLOKI","to":"BEER","amount":715.044453474197481450}
```
---
`GET /exchange?from=USDT&to=GATE&amount=108`

```
--> Status: 200

{"from":"USDT","to":"GATE","amount":15.704803493449786800}
```
---
`GET /exchange?from=BEER&to=FLOKI&amount=1.59`

```
--> Status: 200

{"from":"BEER","to":"FLOKI","amount":0.274018907563025223}
```
---
Failure when ***amount***, ***from*** or ***to*** is empty:

`GET /exchange?from=WBTC&to=USDT&amount=`

```
--> Status: 400
```
---
Failure when ***amount***, ***from*** or ***to*** is missing:

`GET /exchange?from=WBTC&to=USDT`

```
--> Status: 400
```
---
Failure when one of the currencies does not exist in the provided table:

`GET /exchange?from=AAAAA&to=USDT&amount=1.23`

```
--> Status: 400
```
---
Failure when the ***amount*** is a negative number:

`GET /exchange?from=USDT&to=FLOKI&amount=-100`

```
--> Status: 400
```
---
Failure when the ***amount*** is not a number:

`GET /exchange?from=USDT&to=FLOKI&amount=abcd`

```
--> Status: 400
```
---

# TESTING

run:

```
go test -v ./...
```
