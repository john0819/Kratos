package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewUserUsecase, NewSocialUsecase)

// 业务逻辑相关
/*

social - article / comment
user - login / register / profile

user - usercase
article - repo - data

*/
