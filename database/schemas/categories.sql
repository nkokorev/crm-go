create table if not exists categories (
                                        id bigint unsigned auto_increment not null unique primary key,
                                        name varchar(32)
)