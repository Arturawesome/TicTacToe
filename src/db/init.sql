CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY,

    board INTEGER[] NOT NULL,
    current_turn INTEGER NOT NULL,
    game_status INTEGER NOT NULL,

    winner UUID,

    player1_id UUID NOT NULL,
    player1_mark INTEGER NOT NULL,

    player2_id UUID,
    player2_mark INTEGER NOT NULL,

    opponent_type INTEGER NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);