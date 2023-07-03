package resolvers

type VersionResolver struct{}

func (r *VersionResolver) Version() string {
	return "0"
}
