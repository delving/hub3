import {updateRoute} from "./router";

async function doRequest(endpoint, body) {
  // const route = route()
  const response = await fetch('http://localhost:3000/' + endpoint,
    {
      method: 'post',
      headers: {
        'Content-Type': 'application/json;charset=utf-8'
      },
      body: JSON.stringify({
        ...body,
        inventoryId: updateRoute.values.inventoryID
      })
    })
  return await response.json();
}

export async function fetchTree(body) {
  return await doRequest('tree', body)
}

export async function fetchDescription(body) {
  return await doRequest('description', body)
}
