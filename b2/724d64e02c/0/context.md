# Session Context

## User Prompts

### Prompt 1

I want to update the syntaxapi.Searcher interface to add two more methods: SearchFuncs() (iterator.Iterator[Result], error) and SearchTypes() (iterator.Iterator[Result], error). They should be imnplemented the same way Search is implemented, just using different rpcs.

### Prompt 2

remove SearchTypes and coalesce SearchFuncs and SearchTypes into a single SearchNode which will take a enum of Node uint8 -> NodeScope, NodeDefinitionNamespace, NodeReference, NodeDefinitionFunction, NodeDefinitonVar. Propagate changes through the API. Also, add a new method Query(file workspaceapi.URI, query string, captureNames []string) (iterator.Iterator[Result], error) and QueryNode(file workspaceapi.URI, nodeType Node) (iterator.Iterator[Result], error) and implement the proper rpc counter...

### Prompt 3

The Node enum should be a bitflag so we can query multiple nodes in one query. Also let's call it NodeCaptureName.

