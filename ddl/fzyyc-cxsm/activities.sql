create table activities
(
    id                  bigint unsigned auto_increment
        primary key,
    service_provider_id bigint unsigned  default '0'               not null,
    type                varchar(16)                                not null,
    title               varchar(128)                               null,
    content             text                                       null,
    image               varchar(512)                               null,
    link_type           varchar(16)                                null,
    link_value          varchar(256)                               null,
    sort                int unsigned     default '0'               not null,
    status              tinyint unsigned default '1'               not null,
    start_time          datetime                                   null,
    end_time            datetime                                   null,
    created_at          datetime         default CURRENT_TIMESTAMP not null,
    updated_at          datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    comment '平台活动表' collate = utf8mb4_unicode_ci;

create index idx_activities_sort
    on activities (sort);

create index idx_activities_status
    on activities (status);

create index idx_activities_type
    on activities (type);

