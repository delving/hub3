# PathQuery Syntax Documentation

## Overview

The PathQuery syntax is designed to express complex graph traversal queries with optional filtering and value extraction capabilities. This document defines the rules and syntax for creating valid PathQuery strings.

## Syntax Components

A PathQuery string consists of up to three main parts, separated by the pipe delimiter (` | `):

1. **Resource Path** (required): Defines the main traversal path
2. **Filter Path** (optional): Specifies conditions to filter results
3. **Value Path** (optional): Indicates the specific data to extract

### General Syntax

```
[ResourcePath] | [FilterPath]=[FilterValue] | [ValuePath]
```

Multiple filter values for the same filter path can be specified with additional pipe-delimited sections:

```
[ResourcePath] | [FilterPath]=[FilterValue1] | [FilterPath]=[FilterValue2] | [ValuePath]
```

## Path Expression Syntax

### Class-Predicate Notation

Each path segment uses a class-predicate notation:

```
[class]predicate
```

Where:
- `class` is enclosed in square brackets and represents a resource type (e.g., `[crm:E35_Title]`)
- `predicate` follows the class and represents a property or relationship (e.g., `crm:P102_has_title`)

Empty class brackets (`[]`) represent a wildcard for node traversal. For example:
```
[]crm:P102_has_title
```
This indicates traversal via the `crm:P102_has_title` relationship to any node, regardless of its class.

Several classes, separated by `|` without spaces, mean any of them:
```
[]schema:creator->[schema:Person|schema:Organization]schema:name
```
This is the name of a creator that is a person or an organisation.

### Traversal Arrows

The `->` symbol indicates traversal between nodes:

```
[class1]predicate1->[class2]predicate2
```

This indicates traversing from a resource of type `class1` via the `predicate1` relationship to a resource of type `class2`, and then accessing the `predicate2` property.

## Path Types

### Resource Path

The resource path defines the main traversal path through the graph. It must be specified and comes first in the PathQuery string.

Example:
```
[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]crm:P190_has_symbolic_content
```

### Filter Path

The filter path specifies conditions that matching resources must satisfy. It consists of:
- A path expression (similar to resource path)
- An equals sign (`=`)
- A filter value

Example:
```
[crm:E35_Title]crm:P2_has_type=https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h
```

Multiple filter values for the same filter path can be specified with additional pipe-delimited sections:
```
[rdfs:Resource]rdfs:label=Value1 | [rdfs:Resource]rdfs:label=Value2
```

### Value Path

The value path indicates what data should be extracted from matching resources. It has the same syntax as a resource path but appears after any filter expressions.

Example:
```
[crm:E35_Title]crm:P190_has_symbolic_content
```

## Complete Examples

### Simple Path
```
[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title]crm:P190_has_symbolic_content
```

### Path with Filter
```
[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title] | [crm:E35_Title]crm:P2_has_type=https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h
```

### Path with Filter and Value
```
[lrm:F3_Manifestation]crm:P102_has_title->[crm:E35_Title] | [crm:E35_Title]crm:P2_has_type=https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h | [crm:E35_Title]crm:P190_has_symbolic_content
```

### Path with Multiple Values for Same Filter
```
[lrm:F3_Manifestation]crm:P129_is_about | [rdfs:Resource]rdfs:label=Value1 | [rdfs:Resource]rdfs:label=Value2 | [crm:E35_Title]crm:P190_has_symbolic_content
```

## Rules and Constraints

1. The resource path must come first.
2. Filter paths must include a path expression, an equals sign, and a value.
3. All filter expressions must use the same filter path (multiple different filter paths are not supported).
4. Only one value path can be specified.
5. The value path must come after any filter expressions.
6. Empty path strings are valid and will result in a nil PathQuery.

## In the search index

`fragments.ParseResourcePath` reads a resource path in this syntax and turns it into an
Elasticsearch query on the v2 index (`fragments.ResourcePath`). There every linked resource
is a nested `resources` element whose `context` records its route from the record's root:
one element per hop, with its `Level`, the predicate's `SearchLabel` and the `SubjectClass`
of the resource the hop starts from, which is what `[class]predicate` names. So any number
of hops can be queried, and several conditions can hold for the same linked resource:

```
[]schema:creator->[schema:Role]schema:creator->[schema:Person]schema:name   = "Sphinx"
[]schema:creator->[schema:Role]                                            with name "fabrikanten" and creator "Sphinx"
```

A path that ends with a predicate ends at a literal of the last resource; one that ends with
`[class]` ends at that resource, to test its class or its URI. Filters and value paths
(` | `) are not used there: the caller supplies the values.
