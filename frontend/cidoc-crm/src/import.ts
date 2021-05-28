import {classExists, propertyExists} from "./crm";

const modelProperties = {
  "about": true,
  "type": true,
  "properties": true,
  "comment": true,
  "id": true,
  "value": true,
  "if": true,
  "gen": true
}

function propertyPath(ancestors, propertyName, index?: number) {
  let arrayAccess = typeof index === "number" ? `[${index}]` : ""
  return `invalid property value for "${ancestors.join('->')}[${propertyName}]${arrayAccess}":`
}

function checkModel(node, path, timestamp, errors, isCompleteModel?) {
  for (const key of Object.keys(node)) {
    if (!(key in modelProperties)) {
      delete node[key]
    }
  }
  node.uuid = ++timestamp

  if (isCompleteModel) {
    if (node.about !== "#root") {
      errors.push(`${propertyPath(path, 'about')} "${node.about}"`);
    }
  } else if (!propertyExists(node.about)) {
    errors.push(`${propertyPath(path, 'about')} "${node.about}"`);
  }

  if (!Array.isArray(node.type) || node.type.length === 0) {
    errors.push(`${propertyPath(path, 'type')} "${node.type}"`);
  } else {
    let i = 0
    for (const about of node.type) {
      if (!classExists(about)) {
        errors.push(`${propertyPath(path, 'type', i)} "${about}"`);
      }
      i++
    }
  }

  if (node.id !== undefined && typeof node.id !== "string") {
    errors.push(`${propertyPath(path, 'id')} "${node.id}"`);
  }
  if (node.value !== undefined && typeof node.value !== "string") {
    errors.push(`${propertyPath(path, 'value')} "${node.value}"`);
  }
  if (node.if !== undefined && typeof node.if !== "string") {
    errors.push(`${propertyPath(path, 'if')} "${node.if}"`);
  }
  if (node.comment !== undefined && typeof node.comment !== "string") {
    errors.push(`${propertyPath(path, 'comment')} "${node.comment}"`);
  }
  if (node.gen !== false && node.gen !== true && node.gen !== undefined) {
    errors.push(`${propertyPath(path, 'gen')} "${node.gen}"`);
  }

  if (!Array.isArray(node.properties)) {
    errors.push(`${propertyPath(path, 'properties')} "${node.properties}"`);
  } else {
    let i = 0
    for (const property of node.properties) {
      if (!property || typeof property !== "object") {
        errors.push(`${propertyPath(path, 'properties', i)} "${property}"`);
      } else {
        checkModel(property, path.concat(`properties[${i}]`), timestamp, errors)
      }
      i++
    }
  }
}

export async function getModel(isCompleteModel) {
  const json = await navigator.clipboard.readText();
  let model;
  const errors = []
  try {
    model = JSON.parse(json)
    if (model && typeof model === "object" && !Array.isArray(model)) {
      checkModel(model, ["#root"], new Date().getTime(), errors, isCompleteModel)
      model.latest = true
    } else {
      errors.push(`invalid root type`)
    }
  } catch (e) {
    errors.push(e.toString())
  }
  return errors.length === 0 ? [model, null] : [null, errors.join('\n')]
}

