package resolvers

type Root struct {
	VersionResolver
	//ExampleResolver
}

func NewRoot() *Root {
	return &Root{
		VersionResolver: VersionResolver{},
		//ExampleResolver: ExampleResolver{},
	}
}

// func NewExampleResolver() ExampleResolver {
// 	conf := config.MustLoadConfig()
// 	srv := server.MustLoadServer(conf)
//
// 	return ExampleResolver{
// 		srv: srv,
// 	}
// }

type Pagination struct {
	Limit  *int32
	Offset *int32
	Total  *int32
}

type PaginationInput struct {
	Limit  *int32
	Offset *int32
}
