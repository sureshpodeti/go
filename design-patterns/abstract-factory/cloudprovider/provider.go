package cloudprovider

type Storage interface{ Upload() }
type Database interface{ Query() }
type Compute interface{ Run() }
