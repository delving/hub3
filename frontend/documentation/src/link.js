export function linkServiceMethod(service, method) {
  return `#serviceMethod:service=${service.name}&method=${method.name}`
}

export function linkService(service) {
  return `#service:service=${service.name}`
}

export function linkTopic(topicLink) {
  return `#topic:topic=${topicLink.id}`
}