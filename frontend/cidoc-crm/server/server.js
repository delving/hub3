const fs = require('fs');
const jsonServer = require('json-server')
const server = jsonServer.create()
const middlewares = jsonServer.defaults()
const port = process.env.PORT || 3000

server.use(jsonServer.bodyParser)
server.use(middlewares)

server.listen(port, () => {
  console.log('JSON Server is running')
})

server.post('/save', (request, response) => {
  const body = request.body
  console.log(body)
  fs.writeFileSync('models/' + body.filename, JSON.stringify(body))
  response.status(200).jsonp({})
})

server.post('/models', (request, response) => {
  const body = request.body
  console.log(body)
  if(!body.filename) {
    const jsonFiles = fs.readdirSync('models').filter(f => f.endsWith(".json"))
    jsonFiles.sort()
    response.status(200).jsonp(jsonFiles)
  } else {
    const model = fs.readFileSync('models/' + body.filename).toString("utf-8")
    response.status(200).jsonp(JSON.parse(model))
  }
})