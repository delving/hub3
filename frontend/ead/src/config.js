export let config;

let callbacks = []

export function configReady(callback) {
  if (config)
    callback()
  else
    callbacks.push(callback);
}

(async () => {
  const configElement = document.getElementById('delving-config');
  const result = await fetch(configElement.href);
  config = await result.json();
  callbacks.forEach(callback => callback());
  callbacks = [];
})()

