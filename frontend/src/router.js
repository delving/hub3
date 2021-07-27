let callbacks = []
let route = updateRoute()

function updateRoute() {
  const hash = window.location.hash
  const parts = hash.substring(1).split(':')
  const params = {}
  if(parts.length === 2) {
    const properties = parts[1].split("&")
    for (const property of properties) {
      const splitProperty = property.split("=")
      params[splitProperty[0]] = splitProperty.length === 2 ? splitProperty[1] : ''
    }
  }
  console.log(params)

  const event = {name: parts[0], params}
  for (const callback of callbacks) {
    callback({unsubscribe: () => removeListener(callback), ...event})
  }
  return event;
}

window.addEventListener('hashchange', updateRoute)

function removeListener(listenerFunc) {
  callbacks = callbacks.filter(callback => callback !== listenerFunc)
}

export function routeChanged(listenerFunc) {
  callbacks.push(listenerFunc)
  return {unsubscribe: () => removeListener(listenerFunc), ...route}
}