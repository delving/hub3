let urls;

const MISSING = '@missing';

export function configure(config) {
  urls = config.urls;
}

export function linkEad(archive) {
  return createLink(urls.ead, {inventoryID: archive.inventoryID});
}

export function linkCLevel(archive, cLevel) {
  return createLink(urls.cLevel, {
    inventoryID: archive.inventoryID,
    cLevelPath: cLevel.path
  });
}

export function linkEadDescription(archive, search) {
  return createLink(urls.eadDescription, {
    inventoryID: archive.inventoryID,
    query: search.q
  });
}

function getRouteIdFrom(path) {
  for (const [key, url] of Object.entries(urls)) {
    if (url.path.length !== path.length) continue;
    let isMatch = true;
    for (let i = 0; i < path.length; i++) {
      const segment = url.path[i];
      if (segment.indexOf(':') === 0) continue;
      if (segment !== path[i]) {
        isMatch = false;
        break;
      }
    }

    if (isMatch) return key;
  }
}

export function getRouteId() {
  const path = location.pathname.split('/').filter(segment => !!segment)
  return getRouteIdFrom(path);
}

function translate(segment, urlContext) {
  if (segment.indexOf(':') !== -0) return segment;
  const propertyName = segment.substring(1);
  const propertyValue = urlContext[propertyName];
  if (propertyValue === '' || propertyValue === null || propertyValue === undefined) {
    return MISSING;
  }
  return propertyValue;
}

function createLink(link, urlContext) {
  const path = link.path.map(segment => translate(segment, urlContext))
  const query = (link.query || []).map(param => ({
    key: translate(param.key, urlContext),
    value: translate(param.value, urlContext)
  }))
    .filter(param => param.key !== MISSING && param.value !== MISSING)
    .map(param => `${param.key}=${param.value}`);
  let url = `/${path.join('/')}`;
  if (query.length > 0) url += `?${query.join('&')}`;
  return url;
}