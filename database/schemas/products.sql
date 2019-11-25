create table if not exists products (
                                        id bigint unsigned auto_increment not null unique primary key,
                                        name varchar(32)
)