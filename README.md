# ABOUT

Proste REST API slużace do sprawdzenia kursu wymiany jednej waluty na druga, a także przeliczenie wymiany niektórych walut.
Aplikacja pobiera kursy walut z openexchangerates.org API.

---
# DEVELOPMENT

Aby uruchomić aplikację openExchange API wymaga app_id przypisanego do konta użytkownika.  
Założ konto na https://openexchangerates.org/signup/free aby otrzymac app_id.

pobierz repozytorium:

```
git clone git@github.com:i-oles/currency-api.git
cd currency-api
```

nastepnie używając dockera:
```
docker build -t currency-api .
docker run -e APP_ID=<twoje_app_id> -p 8080:8080 currency-api
```

aplikacja domyslnie korzysta z configu znajdującego się w `./config/dev.json`

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

`./example` znajduje się przyklad testowego uderzenia do openExchangeAPI.

# RUNNING

### GET /rates

Returns all possible exchange rate pairs between requested currencies.  
Endpoint wymaga jednego parametru:

- `currencies` - waluty których kurs wymiany nas interesuje.

Wynik jest zwracany w zaokrągleniu do 8 miejsca po przecinku.
W przypadku błędu aplikacja zwraca puste body i statusCode 400.
W przypadku gdy API openExchangeRate zwróci błąd, aplikacja również zwróci status code 400 i puste body.

---
`GET /rates?currencies=GBP,USD`

```json
--> Status: 200

[
    { "from": "USD", "to": "GBP", "rate": 0.74365300 },
    { "from": "GBP", "to": "USD", "rate": 1.34471319 },
]
```
---

`GET /rates?currencies=BTC,INR,LYD`

```json
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
failure when only one param given:

`GET /rates?currencies=BTC`

```json
--> Status: 400
```
---
failure when param ***currencies*** is empty:  

`GET /rates?currencies=`

```json
--> Status: 400
```
---
failure when currency is not found in openExchangeAPI:

`GET /rates?currencies=ABRAKADABRA`

```json
--> Status: 400
```
---



### GET /exchange

Wylicza wartosć wymiany jednej waluty na druga.  

Endpoint wymaga trzech parametrów:
- `from` - waluta krypto, którą chcemy wymienić
- `to` - waluta krypto, która chcemy otrzymać
- `amount` - kwotę krypto jaką chcemy wymienić.

Dane są zwracane na podstawie poniższej tabeli.  
Kolumna decimal placec okresla dokladnosc po przecinku z jaka bedzie zwrocona dana waluta.  
W przypadku błędu aplikacja zwraca puste body i statusCode 400.

| CryptoCurrency | Decimal places | Rate (to USD) |
| ----------- | ----------- | ----------- |
| BEER | 18 | 0.00002461$
| FLOKI | 18 | 0.0001428$
| GATE| 18 | 6.87$
| USDT | 6 | 0.999$
| WBTC | 8 | 57,037.22$

---
`GET /exchange?from=WBTC&to=USDT&amount=1.0`

```json
--> Status: 200

{"from":"WBTC","to":"USDT","amount":57094.314314}
```
---
`GET /exchange?from=GATE&to=WBTC&amount=12.0`

```json
--> Status: 200

{"from":"GATE","to":"WBTC","amount":0.00144537}
```
---
`GET /exchange?from=FLOKI&to=BEER&amount=123.23`

```json
--> Status: 200

{"from":"FLOKI","to":"BEER","amount":715.044453474197481450}
```
---
`GET /exchange?from=USDT&to=GATE&amount=108`

```json
--> Status: 200

{"from":"USDT","to":"GATE","amount":15.704803493449786800}
```
---
`GET /exchange?from=BEER&to=FLOKI&amount=1.59`

```json
--> Status: 200

{"from":"BEER","to":"FLOKI","amount":0.274018907563025223}
```
---
failure when ***amount***, ***from*** or ***to*** is empty:

`GET /exchange?from=WBTC&to=USDT&amount=`

```json
--> Status: 400
```
---
failure when one of three params is missing: 

`GET /exchange?from=WBTC&to=USDT`

```json
--> Status: 400
```
---
failure when one of currencies not exists in given table:

`GET /exchange?from=AAAAA&to=USDT&amount=1.23`

```json
--> Status: 400
```
---
failure when amount is negative number:

`GET /exchange?from=USDT&to=FLOKI&amount=-100`

```json
--> Status: 400
```
---
failure when amount is not a number:

`GET /exchange?from=USDT&to=FLOKI&amount=abcd`

```json
--> Status: 400
```
---
