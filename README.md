<h1 align="center">
DEPRECTATION WARNING:
Following project is currently deprecated, cause of 
	<a href="https://github.com/enteam/enbase">enbase-dotnet</a> implementation. Pull requests and issues are ignored
</h1> 

<p align="center"><img width="30%" src="images/logo.png"/></p>

<h1 align="center">Enbase :leaves:</h1>

<h4 align="center">
  ⚡️ High-availability distributed open source serverless NoSQL realtime database 🐰
</h4>

<p align="center">
  Build powerful realtime cross-platform apps with high-availability backend solution 
</p>

<p align="center">
  
<a href="https://goreportcard.com/report/github.com/enteam/enbase">
  <img src="https://goreportcard.com/badge/github.com/enteam/enbase">
</a>

<a href="https://travis-ci.com/enteam/enbase">
  <img src="https://travis-ci.com/enteam/enbase.svg?branch=master">
</a>

<a href="https://hub.docker.com/r/enteam/enbase/">
  <img src="https://img.shields.io/docker/pulls/enteam/enbase.svg">
</a>

<a href="https://hub.docker.com/r/enteam/enbase/">
  <img src="https://img.shields.io/docker/stars/enteam/enbase.svg">
</a>

<a href="https://github.com/enteam/enbase">
  <img src="https://img.shields.io/github/license/enteam/enbase.svg">
</a>

<a href="https://github.com/enteam/enbase">
  <img src="https://img.shields.io/github/issues/enteam/enbase.svg">
</a>

</p>

|   | Enbase |
| - | ------------ |
| ⚡️ | **Launch your database in minutes** no matter how big is your cluster |
| 📈 | **Highly scalable** from hundreds to tens of thousands of records |
| ✨ | **Reactive** Built-in websocket support  |
| 📱 | **Cross-platform** iOS, Android, and the web |
| ⏱ | **Fast** Powered by Go |
| 🔗 | **NoSQL** - store data like you want |

## Why

Creating reactive apps is the future. You can right now start building your high-performance and highly-avaiable mobile or web app with enbase. Enbase is a solution thats automates your backend functions and packs its into data access layer with powerful access permissions and realtime subscriptions

## :rocket: Quick deployment
#### with docker compose :whale:
```
$ wget https://raw.githubusercontent.com/enteam/enbase/master/docker-compose.yml
$ docker-compose up -d
```
#### with Helm
```
$ helm repo add enteam https://raw.githubusercontent.com/enteam/charts/gh-pages
$ helm install enteam/enbase
```

## ⚡️ Your app setup
```javascript
const Enbase = require('enbase-js-sdk');
const database = new Enbase({
	databaseId: '<enbase-database-id>',
	databaseUrl: 'https://<enbase-address>',
	websocketUrl: 'https://<enbase-address>'
});

// Create new document in 'cats' collection
database.collection('cats').create({
  name: 'Kitty',
  age: 2
})

// Read documents from 'cats' collection, where age is equal to 2
database.collection('cats').read({
  age: 2
})

// Update new document in 'cats' collection
database.collection('cats').update('<id>', {
  age: 3
})

// Delete document in 'cats' collection
database.collection('cats').delete('<id>')

// Stream all documents from 'cats' collection, where age is equal to 2, in realtime
database.collection('cats').stream({
  age: 3
}, (cats) => console.log(cats))
```

## Dashboard

Of course, you can manage your enbase database resources with [enbase-dashboard](https://github.com/enteam/enbase-dashboard). Just access enbase address with your favourite browser and manage create your next awesome app 

<div>
<img style="float:left;" width="48%" src="https://raw.githubusercontent.com/enteam/enbase/master/images/dashboard-screen-1.png">
<img style="float:left;" width="48%" src="https://raw.githubusercontent.com/enteam/enbase/master/images/dashboard-screen-2.png">
</div>

## Contributing

If you have comments, complaints, or ideas for improvements, feel free to open an issue or a pull request!

If you want to contribute, see [Contributing guide](./CONTRIBUTING.md) for details about project setup, testing, etc. If you make a non-trivial contribution, email me, and I'll send you a nice :rabbit: sticker!

## Author and license

**Enbase** was created by [@MatisiekPL](https://github.com/MatisiekPL). Main author and maintainer is [Mateusz Woźniak](https://github.com/MatisiekPL).

**Contributors:** [@MatisiekPL](https://github.com/MatisiekPL), [@Radek-Wawrzyk](https://github.com/Radek-Wawrzyk)

Enbase is available under the MIT license. See the [LICENSE file](./LICENSE) for more info.
