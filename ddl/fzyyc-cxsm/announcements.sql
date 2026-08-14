create table announcements
(
    id                  bigint unsigned auto_increment
        primary key,
    service_provider_id bigint unsigned                            not null,
    title               varchar(128)                               not null,
    content             text                                       null,
    status              tinyint unsigned default '1'               not null,
    created_at          datetime         default CURRENT_TIMESTAMP not null,
    updated_at          datetime         default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint fk_announcements_sp
        foreign key (service_provider_id) references service_providers (id)
            on delete cascade
)
    comment '系统公告表' collate = utf8mb4_unicode_ci;

create index idx_announcements_sp_id
    on announcements (service_provider_id);

create index idx_announcements_status
    on announcements (status);

