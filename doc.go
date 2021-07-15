/*
Package dagcosmos provides a Go implementation of the IPLD DAG-COSMOS spec
(https://github.com/ipld/ipld/tree/master/specs/codecs/dag-cosmos) for
go-ipld-prime (https://github.com/ipld/go-ipld-prime/).
Use the Decode() and Encode() functions directly, or import one of the packages to have their codec
registered into the go-ipld-prime multicodec registry and available from the
cidlink.DefaultLinkSystem.
Nodes encoded with theses codecs _must_ conform to the DAG-COSMOS spec. Specifically,
they should have the non-optional fields shown in the DAG-COSMOS [schemas](https://github.com/ipld/ipld/tree/master/specs/codecs/dag-cosmos):
Use the dagcosmos.Type slab to select the appropriate type (e.g. dagcosmos.Type.Header) for strictness guarantees.
Basic ipld.Nodes will need to have the appropriate fields (and no others) to successfully encode using this codec.
*/
package dagcosmos

//go:generate go run gen.go
//go:generate gofmt -w ipldsch_minima.go ipldsch_satisfaction.go ipldsch_types.go
