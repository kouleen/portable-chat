-- 初始化9位UIN号段
insert or ignore into id_alloc (biz_tag, max_id, step, start_id, end_id)values ('portable-chat', 100000000, 1000, 100000000, 999999999);

