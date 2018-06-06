-- save vars that can change when the server is running
CREATE TABLE public.zglobal_var
(
    zkey text PRIMARY KEY,
    zvalue text DEFAULT ''::text,
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
ALTER TABLE public.zglobal_var OWNER TO vic_user;



--
CREATE TABLE public."user"
(
    id BIGSERIAL, CONSTRAINT userzz_pkey PRIMARY KEY (id),
    username TEXT DEFAULT '' UNIQUE,    
    role TEXT DEFAULT 'ROLE_USER',
    is_suspended BOOLEAN DEFAULT FALSE,
    real_name TEXT DEFAULT '',
    national_id TEXT DEFAULT '',
    phone TEXT DEFAULT '',
    email TEXT DEFAULT '',
    country TEXT DEFAULT '',
    address TEXT DEFAULT '',
    profile_name TEXT DEFAULT '',
    profile_image TEXT DEFAULT '',
    summary TEXT DEFAULT '',
    hashed_password TEXT DEFAULT '',
    login_session TEXT DEFAULT '',
    misc TEXT DEFAULT '',
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
ALTER TABLE public."user" OWNER TO vic_user;
CREATE INDEX userr_i01_username ON public."user" using btree (username);
CREATE INDEX userr_i02_loginsession ON public."user" using btree (login_session);
INSERT INTO public."user" (username, profile_name, role, hashed_password) VALUES ('daominah', 'Dao Min Ah A', 'ROLE_ADMIN', '164f04b29f50874c9330ee60d23a6ff04279c8b21a79afb5721602c6b97e2ac24d7c2070eba5827cab5f3b503bfac26539ec479921c1abadeac4980fcbf3b8a6');
INSERT INTO public."user" (username, profile_name, role, hashed_password) VALUES ('daominah2', 'Dao Min Ah B', 'ROLE_BROADCASTER', '164f04b29f50874c9330ee60d23a6ff04279c8b21a79afb5721602c6b97e2ac24d7c2070eba5827cab5f3b503bfac26539ec479921c1abadeac4980fcbf3b8a6');
INSERT INTO public."user" (username, profile_name, role, hashed_password) VALUES ('daominah3', 'Dao Min Ah U', 'ROLE_USER', '164f04b29f50874c9330ee60d23a6ff04279c8b21a79afb5721602c6b97e2ac24d7c2070eba5827cab5f3b503bfac26539ec479921c1abadeac4980fcbf3b8a6');



--
CREATE TABLE public.user_money
(
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    money_type TEXT DEFAULT '',  -- enum: MT_CASH, MT_EXPERIENCE, MT_ONLINE_DURATION, MT_BROADCAST_DURATION
    val DOUBLE PRECISION DEFAULT 0,
    CONSTRAINT user_money_pkey PRIMARY KEY (user_id, money_type)
);
ALTER TABLE public.user_money OWNER TO vic_user;



--
CREATE TABLE public.user_money_log
(
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
CREATE TABLE public.user_following
(
    user_id_1 BIGINT  DEFAULT 0 REFERENCES public."user" (id),
    user_id_2 BIGINT  DEFAULT 0 REFERENCES public."user" (id),
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT user_following_pkey PRIMARY KEY (user_id_1, user_id_2)
);
ALTER TABLE public.user_following OWNER TO vic_user;
CREATE INDEX user_following_i01 on public.user_following using btree (user_id_2);



--
CREATE TABLE public.user_viewer
(
    user_id BIGINT  DEFAULT 0 REFERENCES public."user" (id),
    viewer_id BIGINT  DEFAULT 0 REFERENCES public."user" (id),
    CONSTRAINT user_viewer_pkey PRIMARY KEY (user_id, viewer_id)
);
ALTER TABLE public.user_viewer OWNER TO vic_user;



--
CREATE TABLE public.user_conversation_moderator
(
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    moderator_id BIGINT  DEFAULT 0 REFERENCES public."user" (id),
    CONSTRAINT user_conversation_moderator_pkey PRIMARY KEY (user_id, moderator_id)
);
ALTER TABLE public.user_conversation_moderator OWNER TO vic_user;



--
CREATE TABLE public.conversation
(
    id BIGSERIAL, CONSTRAINT conversation_pkey PRIMARY KEY (id),
    name TEXT DEFAULT TEXT '',
    conversation_type TEXT DEFAULT 'CONVERSATION_PAIR',  -- enum: CONVERSATION_PAIR, CONVERSATION_GROUP
    pair_key TEXT DEFAULT NULL UNIQUE  -- two users can only have one pair conversation between them, they cant leave the conversation
);
ALTER TABLE public.conversation OWNER TO vic_user;



--
CREATE TABLE public.conversation_member
(
    conversation_id BIGINT DEFAULT 0 REFERENCES conversation (id),
    user_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    CONSTRAINT conversation_member_pkey PRIMARY KEY (conversation_id, user_id),
    is_moderator BOOL DEFAULT FALSE,
    is_blocked BOOL DEFAULT FALSE,
    is_mute BOOL DEFAULT FALSE
);
ALTER TABLE public.conversation_member OWNER TO vic_user;
CREATE INDEX conversation_member_i01 ON public.conversation_member USING btree (user_id);



--
CREATE TABLE public.conversation_message
(
    message_id BIGSERIAL, CONSTRAINT conversation_message_pkey PRIMARY KEY (message_id),
    conversation_id BIGINT DEFAULT 0 REFERENCES conversation (id),
    sender_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    message_content TEXT DEFAULT '',
    created_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
ALTER TABLE public.conversation_message OWNER TO vic_user;
CREATE INDEX conversation_message_i01 ON public.conversation_message USING btree (conversation_id, created_time);


--
CREATE TABLE public.conversation_message_recipient
(
    message_id BIGINT DEFAULT 0 REFERENCES public.conversation_message (message_id),
    recipient_id BIGINT DEFAULT 0 REFERENCES public."user" (id),
    CONSTRAINT conversation_message_recipient_pkey PRIMARY KEY (message_id, recipient_id),
    has_seen BOOL DEFAULT FALSE,
    seen_time TIMESTAMP WITH TIME ZONE DEFAULT '9999-01-01T00:00:00+07:00'
);
ALTER TABLE public.conversation_message_recipient OWNER TO vic_user;