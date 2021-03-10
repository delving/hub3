import {getEndpoint} from "../../gen/clients";

export class API {
  constructor(service) {
    this.service = service;
    this.serviceId = service.name.replace('Service', '');
  }

  getSearchResponseFields() {
    const method = this.service.methods.find(method => method.name === "Search")
    console.log(method.output.fields)
    const hits = method.output.fields.find(field => field.name === "Hits")
    return hits.typeDef.fields
  }

  search(request) {
    return getEndpoint(this.service.name, "Search")(request)
  }

  put(request) {
    return getEndpoint(this.service.name, `Put${this.serviceId}`)(request)
  }

  delete(id) {
    return getEndpoint(this.service.name, `Delete${this.serviceId}`)({id})
  }

  get(id) {
    return getEndpoint(this.service.name, `Get${this.serviceId}`)({id})
  }
}