# stripe-collect

_Collect payments for invoices sent using Stripe_

## intro

**stripe-collect** is tiny application you can self-host in order to enable your clients to pay
your invoices by credit card.

## features

- clients don't have to login to yet an other system
- clients get to pay by credit card in their own currency
- supports different currency for each invoice
- stripe means a great user experience for the recipient and you (receipt email, dashboard)
- not complex, no new terms or processes to learn
- 1 file app you can easily tweak if needed

| Page | Screenshot |
|:---:|:---:|
| Pay | <img src="https://raw.githubusercontent.com/kiasaki/stripe-collect/master/screenshots/pay.png" height="250px" /> |
| Success | <img src="https://raw.githubusercontent.com/kiasaki/stripe-collect/master/screenshots/success.png" height="250px" /> |
| Missing Invoice ID | <img src="https://raw.githubusercontent.com/kiasaki/stripe-collect/master/screenshots/not-found.png" height="250px" /> |

## deploying

And easy way to deploy this app with minimum hassle and maintenance is using **Heroku**.

_**stripe-collect** uses **Google Cloud Storage** as backing store, you'll have to create a bucket
and service account for this app._

Create a new **Heroku** app and set the following environment variables:

```
GOVERSION=go1.7.4
STRIPE_SECRET_KEY=...
GOOGLE_BUCKET_ID=...
GOOGLE_APPLICATION_CREDENTIALS=... (this is the service account json key all on one line)
```

Next and last step is to deploy:

```
git clone git@github.com:kiasaki/stripe-collect.git
cd stripe-collect
heroku git:remote -a <heroku app name>
git push heroku master
```

## developing

Start by create a `.env` file with the following environment variables:

```
export STRIPE_SECRET_KEY='... (make sure stripe's test api key)'
export GOOGLE_BUCKET_ID='...'
export GOOGLE_APPLICATION_CREDENTIALS='... (this one is the json key all on one line)'
```

Basically, there's a `makefile` that `run`s or `build`s the app:

```
$ make
go build -o stripe-collect .
bash -c "source .env; ./stripe-collect"
2016/12/29 10:11:12 started listening on port 3000
```

But if you are developing on the `views/` it's useful to have the server restart on changes. For
this purpose there is a `watch.sh` bash script included:

```
$ ./watch.sh
go build -o stripe-collect .
Path is /[...]/stripe-collect/views
Watching /[...]/stripe-collect/views
2016/12/29 10:11:12 started listening on port 3000
```

## license

MIT. See `license` file.
