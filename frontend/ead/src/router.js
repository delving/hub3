import {config} from './config'
import {parseUrl, queryStore} from "./search/queryStore";

export let route = null;

const MISSING = '@missing';

export function linkEad(archive) {
  return createLink(config.urls.findingAidPage, {inventoryID: archive.inventoryID});
}

export function linkTo(urlId, context) {
  const c = {
    ...route.values,
    baseUrl: config.baseUrl,
    ...context
  };
  console.log(config.urls[urlId], c)
  return createLink(config.urls[urlId], c);
}

export function linkCLevel(archive, cLevel) {
  return createLink(config.urls.findingAidCLevelPage, {
    inventoryID: archive.inventoryID,
    cLevelPath: cLevel.path
  });
}

export function linkEadDescription(archive, search) {
  return createLink(config.urls.findingAidDescriptionPage, {
    inventoryID: archive.inventoryID,
    query: search.q
  });
}

function decodeValue(v) {
  return decodeURIComponent(v.replace('::', '/'));
}

function getRouteFrom(path) {
  for (const component of config.components) {
    for (const route of component.routes) {
      const url = config.urls[route];
      if (url.path.length !== path.length) continue;
      let isMatch = true;
      const values = {};
      for (let i = 0; i < path.length; i++) {
        const segment = url.path[i];
        if (segment.indexOf(':') === 0) {
          values[segment.substring(1)] = decodeValue(path[i]);
        } else if (segment !== path[i]) {
          isMatch = false;
          break;
        }
      }

      if (isMatch) {
        return {
          route,
          config: component,
          values
        };
      }
    }
  }
  return {values: {}};
}

export function updateRoute() {
  const path = location.pathname.split('/').filter(segment => !!segment)
  route = getRouteFrom(path);
  queryStore.parseUrl()
  return route;
}

function translate(segment, urlContext) {
  if (segment.indexOf(':') !== -0) return segment;
  const propertyName = segment.substring(1);
  const propertyValue = urlContext[propertyName];
  if (propertyValue === '' || propertyValue === null || propertyValue === undefined) {
    return MISSING;
  }
  return propertyValue.indexOf('http') === 0
    ? propertyValue
    : propertyValue.replace('/', '::');
}

export function createLink(link, urlContext) {
  const path = link.path.map(segment => translate(segment, urlContext))
  const query = (link.query || []).map(param => ({
    key: translate(param.key, urlContext),
    value: translate(param.value, urlContext),
    prefix: param.prefix
  }))
    .filter(param => param.key !== MISSING && param.value !== MISSING)
    .map(param => `${param.key}=${param.prefix || ''}${param.value}`);
  const joinedPath = path.join('/');
  let url = joinedPath.indexOf('http') !== 0 ? `/${joinedPath}` : joinedPath;
  if (query.length > 0) url += `?${query.join('&')}`;
  return url;
}