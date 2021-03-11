import {getEndpoint} from "../../gen/clients";

export class API {
  constructor(service) {
    this.service = service;
    this.domainObjectName = service.name.replace('Service', '');
  }

  getSearchResponseFields() {
    const method = this.service.methods.find(method => method.name === "Search")
    console.log(method.output.fields)
    const hits = method.output.fields.find(field => field.name === "Hits")
    return hits.typeDef ? hits.typeDef.fields : [{name: 'uuid'}, {name: 'prefix'}]
  }

  search(request) {
    return getEndpoint(this.service.name, "Search")(request)
  }

  put(domainObject) {
    console.log(domainObject)
    return getEndpoint(this.service.name, `Put${this.domainObjectName}`)(domainObject)
  }

  delete(id) {
    return getEndpoint(this.service.name, `Delete${this.domainObjectName}`)({id})
  }

  get(id) {
    return getEndpoint(this.service.name, `Get${this.domainObjectName}`)({id})
  }
}