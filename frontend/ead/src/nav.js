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