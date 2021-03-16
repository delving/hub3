const api = require("./api");
const fs = require('fs');
const parseEad = require("./ead-parser");

const cache = {
  "4.OSK": parseEad(fs.readFileSync("./4.OSK.xml")),
  // "1.04.02": parseEad(fs.readFileSync("./1.04.02.xml"))
};

const jsonServer = require('json-server')
const server = jsonServer.create()
const middlewares = jsonServer.defaults()
const port = process.env.PORT || 3000

server.use(jsonServer.bodyParser)
server.use(middlewares)

server.listen(port, () => {
  console.log('JSON Server is running')
})

server.post('/tree', (request, response) => {
  const ead = cache[request.body.inventoryId]
  const data = api.fetchTree(ead, request.body)
  response.status(200).jsonp(data)
})

server.post('/description', (request, response) => {
  const ead = cache[request.body.inventoryId]
  const data = api.fetchDescription(ead, request.body)
  response.status(200).jsonp(data)
})