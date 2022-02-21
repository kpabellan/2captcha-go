# Golang Module for 2Captcha API
The easiest way to quickly integrate [2Captcha] into your code.

- [Installation](#installation)
- [Configuration](#configuration)
- [Solve captcha](#solve-captcha)
  - [ReCaptcha v2](#recaptcha-v2)
  - [ReCaptcha v3](#recaptcha-v3)
- [Other methods](#other-methods)
  - [balance](#balance)


## Installation
To install the api client, use this:

```bash
go get -u github.com/kpabellan/2captcha-go
```

## Configuration

Import the module like this:

```go
import (
        "github.com/kpabellan/2captcha-go"
)
```

`Client` instance can be created like this:

```go
client := api2captcha.NewClient("YOUR_API_KEY")
```

There are few options that can be configured:

```go
client.RecaptchaTimeout = 600
client.PollingInterval = 10
```

### Client instance options

|Option|Default value|Description|
|---|---|---|
|recaptcha_timeout|600|Timeout for ReCaptcha in seconds. Defines how long the module tries to get the answer from `res.php` API endpoint|
|polling_interval|10|Interval in seconds between requests to `res.php` API endpoint, setting values less than 5 seconds is not recommended|

### ReCaptcha v2
Use this method to solve ReCaptcha V2 and obtain a token to bypass the protection.

```go
cap := api2captcha.ReCaptcha{
   SiteKey: "6Le-wvkSVVABCPBMRTvw0Q4Muexq1bi0DJwx_mJ-",
   Url: "https://mysite.com/page/with/recaptcha",
   Invisible: true,
   Action: "verify",
}
req := cap.ToRequest()
req.SetProxy("HTTPS", "login:password@IP_address:PORT")
code, err := client.solve(req)
```

### ReCaptcha v3
This method provides ReCaptcha V3 solver and returns a token.

```go
cap := api2captcha.ReCaptcha{
   SiteKey: "6Le-wvkSVVABCPBMRTvw0Q4Muexq1bi0DJwx_mJ-",
   Url: "https://mysite.com/page/with/recaptcha",
   Version: "v3",
   Action: "verify",
   Score: 0.3,
}
req := cap.ToRequest()
req.SetProxy("HTTPS", "login:password@IP_address:PORT")
code, err := client.solve(req)
```

## Other methods

### Balance
Use this method to get your account's balance

```go
balance, err := client.GetBalance()
if err != nil {
   log.Fatal(err);
}
```