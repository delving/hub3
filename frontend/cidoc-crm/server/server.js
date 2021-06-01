const fs = require('fs');
const http = require('http')
const path = require("path");

const port = process.env.PORT ? +process.env.PORT : 3000
http.createServer(function (req, res) {
  console.log(`${req.method} ${req.url}`)
  res.setHeader("Access-Control-Allow-Origin", "*")
  res.setHeader("Access-Control-Allow-Headers", "*")
  if(req.method === 'OPTIONS') {
    res.writeHead(200)
    res.end()
    return
  }
  let data = '';
  if(req.method === 'POST') {
    req.on('data', chunk => {
      data += chunk;
    })
    req.on('end', () => {
      if (req.url === '/save') {
        save(JSON.parse(data), res)
      } else if (req.url === '/models') {
        models(JSON.parse(data), res)
      }
    })
  }
  if (req.method === 'GET') {
    let url = req.url === '/' ? '/index.html' : req.url
    const p = path.join('public', ...url.substring(1).split('/'))
    const file = fs.readFileSync(p)
    if (url.endsWith('.svg')) {
      res.setHeader('Content-Type', 'image/svg+xml')
    } else  if (url.endsWith('.html')) {
      res.setHeader('Content-Type', 'text/html')
    } else  if (url.endsWith('.css')) {
      res.setHeader('Content-Type', 'text/css')
    }
    res.writeHead(200)
    res.end(file)
  }
}).listen(port); //the server object listens on port 8080
console.log(`Listening on port ${port}`)

function save(body, response) {
  console.log(body)
  fs.writeFileSync(path.join('models', body.filename), JSON.stringify(body))
  response.setHeader('Content-Type', 'application/json')
  response.writeHead(200)
  response.end('{}')
}

function models(body, response) {
  console.log(body)
  response.setHeader('Content-Type', 'application/json')
  if (!body.filename) {
    const jsonFiles = fs.readdirSync('models').filter(f => f.endsWith(".json"))
    jsonFiles.sort()
    response.writeHead(200)
    response.end(JSON.stringify(jsonFiles))
  } else {
    const model = fs.readFileSync(path.join('models', body.filename)).toString("utf-8")
    response.writeHead(200)
    response.end(model)
  }
}