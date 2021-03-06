-- save vars that can change when the server is running
CREATE TABLE public.zglobal_var (
    zkey text PRIMARY KEY,
    zvalue text DEFAULT ''::text,
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
ALTER TABLE public.zglobal_var OWNER TO vic_user;



--
CREATE TABLE public."user" (
    id BIGSERIAL, CONSTRAINT userzz_pkey PRIMARY KEY (id),
    username TEXT DEFAULT '' UNIQUE,    
    role TEXT DEFAULT 'ROLE_USER',
    is_suspended BOOLEAN DEFAULT FALSE,
    real_name TEXT DEFAULT '',
    national_id TEXT DEFAULT '',
    sex TEXT DEFAULT 'SEX_UNDEFINED',
    phone TEXT DEFAULT '',
    email TEXT DEFAULT '',
    country TEXT DEFAULT '',
    address TEXT DEFAULT '',
    profile_name TEXT DEFAULT '',
    profile_image TEXT DEFAULT '',
    summary TEXT DEFAULT 'No information given',
    hashed_password TEXT DEFAULT '164f04b29f50874c9330ee60d23a6ff04279c8b21a79afb5721602c6b97e2ac24d7c2070eba5827cab5f3b503bfac26539ec479921c1abadeac4980fcbf3b8a6',
    login_session TEXT DEFAULT 'hohohaha',
    misc TEXT DEFAULT '{}',
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
ALTER TABLE public."user" OWNER TO vic_user;
CREATE INDEX userr_i01_username ON public."user" using btree (username);
CREATE INDEX userr_i02_loginsession ON public."user" using btree (login_session);
CREATE INDEX userr_i03 ON public."user" using btree (profile_name);
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('daominah', 'Dao Min Ah A1', 'ROLE_ADMIN');
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('daominah2', 'Dao Min Ah B2', 'ROLE_BROADCASTER');
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('daominah3', 'Dao Min Ah U3', 'ROLE_USER');
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('daominah4', 'Dao Min Ah B4', 'ROLE_BROADCASTER');
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('daominah5', 'Dao Min Ah U5', 'ROLE_USER');
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('daominah6', 'Dao Min Ah U6', 'ROLE_USER');
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('daominah7', 'Dao Min Ah U7', 'ROLE_USER');
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('tungdt', 'Tung', 'ROLE_USER');
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('tungdt2', 'Tùng', 'ROLE_USER');
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('tungdt3', 'Đào Thanh Tùng', 'ROLE_USER');    
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('landt', 'Lán', 'ROLE_USER');
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('vantt', 'Vân', 'ROLE_USER');
INSERT INTO public."user" (username, profile_name, role)
    VALUES ('tungdt4', '9 test search', 'ROLE_USER');

    
    

--
CREATE TABLE public.user_money (
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    money_type TEXT DEFAULT '',
        -- enum: MT_CASH, MT_EXPERIENCE, MT_ONLINE_DURATION, MT_BROADCAST_DURATION
    val DOUBLE PRECISION DEFAULT 0,
    CONSTRAINT user_money_pkey PRIMARY KEY (user_id, money_type)
);
ALTER TABLE public.user_money OWNER TO vic_user;



--
CREATE TABLE public.user_money_log (
    id BIGSERIAL, CONSTRAINT user_money_log_pkey PRIMARY KEY (id),
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    money_type TEXT DEFAULT '',
    changed_val DOUBLE PRECISION DEFAULT 0,
    money_before DOUBLE PRECISION DEFAULT 0,
    money_after DOUBLE PRECISION DEFAULT 0,
    reason TEXT DEFAULT '',
    misc TEXT DEFAULT ''
);
ALTER TABLE public.user_money_log OWNER TO vic_user;
CREATE INDEX user_money_log_i01_user_time on public.user_money_log using btree
    (user_id, created_time);
    
    

--
CREATE TABLE public.user_following (
    user_id_1 BIGINT  DEFAULT 0 REFERENCES public."user" (id),
    user_id_2 BIGINT  DEFAULT 0 REFERENCES public."user" (id),
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT user_following_pkey PRIMARY KEY (user_id_1, user_id_2)
);
ALTER TABLE public.user_following OWNER TO vic_user;
CREATE INDEX user_following_i01 on public.user_following using btree (user_id_2);



--
CREATE TABLE public.user_viewer (
    user_id BIGINT  DEFAULT 0 REFERENCES public."user" (id),
    viewer_id BIGINT  DEFAULT 0 REFERENCES public."user" (id),
    CONSTRAINT user_viewer_pkey PRIMARY KEY (user_id, viewer_id)
);
ALTER TABLE public.user_viewer OWNER TO vic_user;



--
CREATE TABLE public.user_conversation_moderator (
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    moderator_id BIGINT  DEFAULT 0 REFERENCES public."user" (id),
    CONSTRAINT user_conversation_moderator_pkey PRIMARY KEY (user_id, moderator_id)
);
ALTER TABLE public.user_conversation_moderator OWNER TO vic_user;




--
CREATE TABLE public.user_login (
    id BIGSERIAL, CONSTRAINT user_login_pkey PRIMARY KEY (id),
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    login_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    logout_time TIMESTAMP WITH TIME ZONE DEFAULT '9999-01-01T00:00:00+07:00',
    network_address TEXT DEFAULT '',
    device_name TEXT DEFAULT '',
    app_name TEXT DEFAULT ''
);
CREATE INDEX user_login_i01 ON public.user_login USING btree
    (user_id, login_time);





--
CREATE TABLE public.team (
    team_id BIGSERIAL, CONSTRAINT team_pkey PRIMARY KEY (team_id),
    team_name TEXT DEFAULT '' UNIQUE,
    team_image TEXT DEFAULT '',
    summary TEXT DEFAULT '',
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);



--
CREATE TABLE public.team_member (
    team_id BIGINT DEFAULT 0 REFERENCES public.team (team_id),
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    CONSTRAINT team_member_pkey PRIMARY KEY (team_id, user_id),
    is_captain BOOL DEFAULT FALSE,
    joined_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE UNIQUE INDEX team_member_i01 ON public.team_member USING btree
    (team_id, is_captain) WHERE is_captain = TRUE;
CREATE UNIQUE INDEX team_member_i02 ON public.team_member USING btree
    (user_id);
    
    
    
--
CREATE TABLE public.team_joining_request (
    team_id BIGINT DEFAULT 0 REFERENCES public.team (team_id),
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    CONSTRAINT team_joining_request_pkey PRIMARY KEY (team_id, user_id),
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);




--
CREATE TABLE public.conversation (
    id BIGSERIAL, CONSTRAINT conversation_pkey PRIMARY KEY (id),
    name TEXT DEFAULT TEXT '',
    conversation_type TEXT DEFAULT 'CONVERSATION_PAIR',
        -- enum: CONVERSATION_PAIR, CONVERSATION_GROUP
    pair_key TEXT DEFAULT '' 
);
ALTER TABLE public.conversation OWNER TO vic_user;
-- two users can only have one pair conversation between them
CREATE UNIQUE INDEX conversation_i01 ON public.conversation USING btree
    (pair_key) WHERE pair_key != '';


--
CREATE TABLE public.conversation_member (
    conversation_id BIGINT DEFAULT 0 REFERENCES conversation (id),
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    CONSTRAINT conversation_member_pkey PRIMARY KEY (conversation_id, user_id),
    is_moderator BOOL DEFAULT FALSE,
    is_blocked BOOL DEFAULT FALSE,
    is_mute BOOL DEFAULT FALSE
);
ALTER TABLE public.conversation_member OWNER TO vic_user;
CREATE INDEX conv_member_i01 ON public.conversation_member USING btree (user_id);



--
CREATE TABLE public.conversation_message (
    message_id BIGSERIAL, CONSTRAINT conversation_message_pkey PRIMARY KEY (message_id),
    conversation_id BIGINT DEFAULT 0 REFERENCES conversation (id),
    sender_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    message_content TEXT DEFAULT '',
    display_type TEXT DEFAULT 'DISPLAY_TYPE_NORMAL',
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
ALTER TABLE public.conversation_message OWNER TO vic_user;
CREATE INDEX conversation_message_i01 ON public.conversation_message USING btree
    (conversation_id, created_time);


--
CREATE TABLE public.conversation_message_recipient (
    message_id BIGINT DEFAULT 0 REFERENCES public.conversation_message (message_id),
    recipient_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    CONSTRAINT conversation_message_recipient_pkey PRIMARY KEY (message_id, recipient_id),
    has_seen BOOL DEFAULT FALSE,
    seen_time TIMESTAMP WITH TIME ZONE DEFAULT '9999-01-01T00:00:00+07:00'
);
ALTER TABLE public.conversation_message_recipient OWNER TO vic_user;



--
CREATE TABLE public.conversation_cheer (
    id BIGSERIAL, CONSTRAINT cheer_pkey PRIMARY KEY (id),
    conversation_id BIGINT DEFAULT 0 REFERENCES public.conversation (id),
    cheerer_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    target_user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    cheer_type TEXT DEFAULT '',  -- CHEER_FOR_USER, CHEER_FOR_TEAM
    team_id BIGINT DEFAULT 0 REFERENCES public.team (team_id),
    val DOUBLE PRECISION DEFAULT 0,
    cheer_message TEXT DEFAULT '',
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    misc TEXT DEFAULT ''
);
CREATE INDEX conversation_cheer_i01 ON public.conversation_cheer USING btree
    (cheerer_id, created_time);
CREATE INDEX conversation_cheer_i02 ON public.conversation_cheer USING btree
    (target_user_id, created_time);
CREATE INDEX conversation_cheer_i03 ON public.conversation_cheer USING btree
    (team_id, created_time);




--
CREATE TABLE public.rank (
    rank_id BIGSERIAL, CONSTRAINT rank_pkey PRIMARY KEY (rank_id),
    rank_name TEXT DEFAULT '' UNIQUE,
    started_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
INSERT INTO public.rank (rank_name) VALUES ('Test1');
INSERT INTO public.rank (rank_name) VALUES ('Test2');
INSERT INTO public.rank (rank_name) VALUES ('Received cash this day');
INSERT INTO public.rank (rank_name) VALUES ('Received cash this week');
INSERT INTO public.rank (rank_name) VALUES ('Received cash this month');
INSERT INTO public.rank (rank_name) VALUES ('Received cash all time');
INSERT INTO public.rank (rank_name) VALUES ('Sent cash this day');
INSERT INTO public.rank (rank_name) VALUES ('Sent cash this week');
INSERT INTO public.rank (rank_name) VALUES ('Sent cash this month');
INSERT INTO public.rank (rank_name) VALUES ('Sent cash all time');
INSERT INTO public.rank (rank_name) VALUES ('Purchased cash this day');
INSERT INTO public.rank (rank_name) VALUES ('Purchased cash this week');
INSERT INTO public.rank (rank_name) VALUES ('Purchased cash this month');
INSERT INTO public.rank (rank_name) VALUES ('Purchased cash all time');
INSERT INTO public.rank (rank_name) VALUES ('Number of followers this week');
INSERT INTO public.rank (rank_name) VALUES ('Number of followers all time');



--
CREATE TABLE public.rank_key (
    rank_id BIGINT DEFAULT 0 REFERENCES public.rank (rank_id),
    user_id BIGINT DEFAULT 0,
    CONSTRAINT rank_key_pkey PRIMARY KEY (rank_id, user_id),
    rkey DOUBLE PRECISION DEFAULT 0,
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX rank_key_i01 ON public.rank_key USING btree
    (rank_id, rkey, last_modified, user_id);


--
CREATE TABLE public.rank_archive (
    archive_id BIGSERIAL, CONSTRAINT rank_archive_pkey PRIMARY KEY (archive_id),
    rank_id BIGINT DEFAULT 0,
    rank_name TEXT DEFAULT '',
    started_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    finished_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    top_10 TEXT DEFAULT '[]',
    full_order TEXT DEFAULT '[]'
);

CREATE TABLE public.gift (
    id BIGSERIAL, CONSTRAINT gift_pkey PRIMARY KEY (id),
    name TEXT DEFAULT '' UNIQUE,
    val DOUBLE PRECISION DEFAULT 0,
    image TEXT DEFAULT '',
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
INSERT INTO public.gift (name, val) VALUES ('Candy', 10000);
INSERT INTO public.gift (name, val) VALUES ('Banana', 20000);
INSERT INTO public.gift (name, val) VALUES ('Tortoise', 30000);
INSERT INTO public.gift (name, val) VALUES ('Four-leaf Clover', 50000);
INSERT INTO public.gift (name, val) VALUES ('Firework', 100000);
INSERT INTO public.gift (name, val) VALUES ('Champagne', 200000);
INSERT INTO public.gift (name, val) VALUES ('Cigar', 500000);
INSERT INTO public.gift (name, val) VALUES ('Sports car ', 1000000);

CREATE TABLE public.stream_archive (
    id BIGSERIAL, CONSTRAINT stream_archive_pkey PRIMARY KEY (id),
    broadcaster_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    stream_image TEXT DEFAULT '',
    stream_name TEXT DEFAULT '',
    started_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    finished_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    n_viewers BIGINT DEFAULT 0,
    n_reports BIGINT DEFAULT 0,
    viewers TEXT DEFAULT '{}',
    reports TEXT DEFAULT '{}',
    conversation_id BIGINT DEFAULT 0 REFERENCES public.conversation (id)
);
CREATE INDEX stream_archive_i01 ON public.stream_archive
    using btree (broadcaster_id, started_time);

    
    
CREATE TABLE match_single (
    id TEXT, CONSTRAINT match_single_pkey PRIMARY KEY (id),
    game_code TEXT DEFAULT '',
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    started_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    money_type TEXT DEFAULT '',
    base_money DOUBLE PRECISION DEFAULT 0,
    result_changed_money DOUBLE PRECISION DEFAULT 0,
    result_detail       TEXT DEFAULT ''
);
CREATE INDEX match_single_i01 ON public.match_single
    USING btree (game_code, user_id, started_time);
    
CREATE TABLE match_multi (
    id TEXT, CONSTRAINT match_multi_pkey PRIMARY KEY (id),
    game_code TEXT DEFAULT '',
    started_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    money_type TEXT DEFAULT '',
    base_money DOUBLE PRECISION DEFAULT 0,
    result_changed_money DOUBLE PRECISION DEFAULT 0,
    result_detail TEXT DEFAULT ''
);
CREATE INDEX match_multi_i01 ON public.match_multi
    USING btree (game_code, started_time);
    
CREATE TABLE match_multi_participant (
    match_id TEXT DEFAULT '' REFERENCES public.match_multi (id),
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    result_changed_money DOUBLE PRECISION DEFAULT 0,
    CONSTRAINT match_multi_participant_pkey PRIMARY KEY (match_id, user_id)
);





CREATE TABLE finance_charge (
    id BIGSERIAL, CONSTRAINT finance_charge_pkey PRIMARY KEY (id),
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    charging_type TEXT DEFAULT '',
    -- 
    http_request TEXT DEFAULT '',
    charging_input TEXT DEFAULT '{}',
    --
    http_response TEXT DEFAULT '',
    vnd_value DOUBLE PRECISION DEFAULT 0,
    transaction_id_3rd_party TEXT DEFAULT '',
    is_successful BOOL DEFAULT FALSE,
    error_message TEXT DEFAULT '',
    in_app_value DOUBLE PRECISION DEFAULT 0,
    money_log_id BIGINT DEFAULT 0,
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX finance_charge_i01 ON public.finance_charge (created_time);
CREATE INDEX finance_charge_i02 ON public.finance_charge (user_id, created_time);
CREATE INDEX finance_charge_i03 ON public.finance_charge (money_log_id);

    
    
    
CREATE TABLE finance_withdraw (
    id BIGSERIAL, CONSTRAINT finance_withdraw_pkey PRIMARY KEY (id),
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    withdrawing_type TEXT DEFAULT '',
    --
    in_app_value DOUBLE PRECISION DEFAULT 0,
    vnd_value DOUBLE PRECISION DEFAULT 0,
    money_log_id BIGINT DEFAULT 0,
    is_denied_by_admin BOOL DEFAULT FALSE,
    denied_reason TEXT DEFAULT '',
    --
    http_request TEXT DEFAULT '',
    http_response TEXT DEFAULT '',
    transaction_id_3rd_party TEXT DEFAULT '',
    is_successful BOOL DEFAULT FALSE,
    error_message TEXT DEFAULT '',
    withdrawing_output TEXT DEFAULT '{}',
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX finance_withdraw_i01 ON public.finance_withdraw (created_time);
CREATE INDEX finance_withdraw_i02 ON public.finance_withdraw (user_id, created_time);
CREATE INDEX finance_withdraw_i03 ON public.finance_withdraw (money_log_id);



CREATE TABLE video (
    id BIGSERIAL, CONSTRAINT video_pkey PRIMARY KEY (id),
    name TEXT DEFAULT '',
    cate_id BIGINT DEFAULT 0,
    image TEXT DEFAULT '',
    video TEXT DEFAULT '',
    price DOUBLE PRECISION DEFAULT 0,
    description TEXT DEFAULT '',
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
ALTER TABLE video ADD is_hot BOOLEAN DEFAULT FALSE;


CREATE TABLE video_categories (
  id BIGSERIAL, CONSTRAINT video_categories_pkey PRIMARY KEY (id),
  name TEXT DEFAULT '',
  description TEXT DEFAULT '',
  created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


CREATE TABLE video_buyer (
    video_id BIGINT DEFAULT 0 REFERENCES video (id),
    user_id BIGINT DEFAULT 0 REFERENCES "user" (id),
    bought_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT video_buyer_pkey PRIMARY KEY (video_id, user_id)
);

CREATE TABLE video_categories_buyer (
    category_id BIGINT DEFAULT 0 REFERENCES video_categories (id),
    user_id BIGINT DEFAULT 0 REFERENCES "user" (id),
    bought_date TEXT DEFAULT '',
    CONSTRAINT video_categories_buyer_pkey PRIMARY KEY (category_id, user_id, bought_date)
);

CREATE TABLE ads
(
  id BIGSERIAL, CONSTRAINT ads_pkey PRIMARY KEY (id),
  name TEXT DEFAULT '',
  url TEXT DEFAULT '',
  image TEXT DEFAULT '',
  "type" TEXT DEFAULT '',
  description TEXT DEFAULT '',
  created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);