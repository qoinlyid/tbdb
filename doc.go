/*
Package tbdb is TigerBeetle DB client wrapper with enchanced methods. This package can be as standalone usage or as a Qore dependency.

# Qore dependency usage

Instance should embed on Qore module:

	type Module struct {
		Tbdb *tbdb.Instance
	}

	func InitModule() *Module {
		m := &Module{
			Tbdb: tbdb.New()
		}

		return m
	}

Then, Open & Close instance & connection will be handled automatically by Qore.

# Standalone

You can use this package as a standalone Go package to suit your needs. You can place the Instance pointer as a singleton anywhere you like.

	instance := tbdb.New()
	if err := instance.Open(); err != nil {
		log.Fatal(err)
	}
	defer func() { _ = instance.Close() }()

The only thing that you must be aware is, the Open-Close should be in that order.
*/
package tbdb
