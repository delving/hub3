// @ts-ignore
import model from '../../../ikuzo/cidoc-crm/model.json'

function defineSuperTypes(type, superTypes) {
  if (!superTypes) return;
  for (const superType of superTypes) {
    const parentClass = model.classes.find(type => type.about === superType.resource)
    type.superClasses.push(parentClass)
    if(!parentClass.subClasses.find(subClass => subClass === type)) {
      parentClass.subClasses.push(type)
    }
    defineSuperTypes(type, parentClass.subClassOf)
  }
}

function defineSuperProperties(type, superProperties) {
  if (!superProperties) return;
  for (const superProperty of superProperties) {
    const parentProperty = model.properties.find(type => type.about === superProperty.resource)
    type.superProperties.push(parentProperty)
    defineSuperProperties(type, parentProperty.subPropertyOf)
  }
}

function extractLabels(item) {
  for(const label of item.labels) {
    item.labels[label.lang] = label.text
  }
}

for (const type of model.classes) {
  type.properties = []
  type.superClasses = []
  type.subClasses = []
}

for (const type of model.classes) {
  extractLabels(type)
  defineSuperTypes(type, type.subClassOf)
}

for(const property of model.properties) {
  property.superProperties = []
  property.dotOnes = []
}

for (const property of model.properties) {
  extractLabels(property)
  defineSuperProperties(property, property.subPropertyOf)
  if(property.about.indexOf(".1_") === -1) {
    const allowedInType = model.classes.find(type => type.about === property.domain[0].resource)
    allowedInType.properties.push(property)
  } else {
    for (const domain of property.domain) {
      const allowedInProperty = model.properties.find(type => type.about === domain.resource)
      allowedInProperty.dotOnes.push(property)
    }
  }
  property.type = model.classes.find(type => type.about === property.range.resource);
}

export function getAllowedProperties(about, typeIds) {
  const allowedProperties = new Set()
  if(about !== '#root') {
    const propertyContext = model.properties.find(p => p.about === about)
    propertyContext.dotOnes.forEach(dotOne => allowedProperties.add(dotOne))
  }

  for (const typeId of typeIds) {
    const type = model.classes.find(type => type.about === typeId)
    type.properties.forEach(p => allowedProperties.add(p))
    for (const superClass of type.superClasses) {
      for (const property of superClass.properties) {
        allowedProperties.add(property)
      }
    }
  }
  const asArray = Array.from(allowedProperties)
  asArray.sort(compare)
  return asArray
}

export function getAllowedTypes(propertyAbout, noRestrictions) {
  if(noRestrictions) {
    return [...model.classes]
  }

  const allowedTypes = new Set()
  const p = getProperty(propertyAbout)
  const allProperties = [p, ...p.superProperties]
  for(const property of allProperties) {
    const range = model.classes.find(type => type.about === property.range.resource);
    allowedTypes.add(range)
    range.subClasses.forEach(s => allowedTypes.add(s))
  }
  const asArray = Array.from(allowedTypes)
  asArray.sort(compare)
  return asArray
}

export function getProperty(propertyAbout) {
  return model.properties.find(p => p.about === propertyAbout);
}

model.classes.push({
  about: "http://www.w3.org/2000/01/rdf-schema#Literal",
  subClassOf: null,
  properties: [],
  labels: {en: "#Literal"},
  subClasses: [],
  superClasses: []
})

model.classes.push({
  about: "http://www.w3.org/2001/XMLSchema#dateTime",
  subClassOf: null,
  properties: [],
  labels: {en: "#Datetime"},
  subClasses: [],
  superClasses: []
})

model.classes.sort(compare)
model.properties.sort(compare)

for (const type of model.classes) {
  type.subClasses.sort(compare)
}

function aboutToNumber(i) {
  const parts = i.about.split("_");
  if(parts.length === 1) return -1;
  const hasI = parts[0].indexOf('i') >= 0
  const numericValue = +parts[0].substring(1).replace('i', '') * 10
  const valueWithI = hasI ? numericValue + 1 : numericValue
  return valueWithI
}

export function compare(a, b) {
  return aboutToNumber(a) - aboutToNumber(b)
}

console.log(model)

export const crm = model
