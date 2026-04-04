package seed

import (
	"time"

	"retro-treasure-backend/internal/model"
	"retro-treasure-backend/internal/repository"
)

func Load(repo *repository.MemoryRepository) {
	repo.SeedAreas([]model.Area{{ID: 1, Name: "草原の入口", Description: "初心者向けの探索エリア。", RequiredLevel: 1, StaminaCost: 1, IsActive: true, SortOrder: 1}, {ID: 2, Name: "古代遺跡", Description: "希少な欠片が眠る遺跡。", RequiredLevel: 3, StaminaCost: 2, IsActive: true, SortOrder: 2}, {ID: 3, Name: "深緑の森", Description: "少し危険だが報酬が良い。", RequiredLevel: 5, StaminaCost: 3, IsActive: true, SortOrder: 3}})
	repo.SeedItems([]model.Item{{ID: 1, Name: "きらめく石", Description: "淡く輝く石。", Rarity: 1, ItemType: "treasure", SellPrice: 10, IsEncyclopediaTarget: true}, {ID: 2, Name: "古びたコイン", Description: "昔の時代の硬貨。", Rarity: 1, ItemType: "treasure", SellPrice: 12, IsEncyclopediaTarget: true}, {ID: 3, Name: "幻の羽根", Description: "めったに見つからない羽根。", Rarity: 3, ItemType: "treasure", SellPrice: 80, IsEncyclopediaTarget: true}, {ID: 4, Name: "遺跡の欠片", Description: "遺跡の一部と思われる破片。", Rarity: 2, ItemType: "material", SellPrice: 30, IsEncyclopediaTarget: true}, {ID: 5, Name: "神秘の宝玉", Description: "強い光を宿した宝玉。", Rarity: 4, ItemType: "treasure", SellPrice: 150, IsEncyclopediaTarget: true}})
	repo.SeedCards([]model.CharacterCard{
		{ID: 1, Name: "見習いトレジャーハンター", Description: "前衛で戦うバランス型。", Rarity: 1, Element: "fire", MaxHP: 24, Attack: 8, Defense: 3, IsStarter: true, PortraitLabel: "TH", FrameStyle: "bronze"},
		{ID: 2, Name: "ランタンメイジ", Description: "灯火の魔法で援護する。", Rarity: 1, Element: "light", MaxHP: 20, Attack: 9, Defense: 2, IsStarter: true, PortraitLabel: "LM", FrameStyle: "bronze"},
		{ID: 3, Name: "レンジャースカウト", Description: "素早い弓手。", Rarity: 1, Element: "wind", MaxHP: 18, Attack: 10, Defense: 2, IsStarter: true, PortraitLabel: "RS", FrameStyle: "bronze"},
		{ID: 4, Name: "シールドベアラー", Description: "盾で味方を守る。", Rarity: 1, Element: "earth", MaxHP: 28, Attack: 6, Defense: 5, IsStarter: true, PortraitLabel: "SB", FrameStyle: "bronze"},
		{ID: 5, Name: "ツインダガー", Description: "手数で押し切る。", Rarity: 1, Element: "dark", MaxHP: 19, Attack: 9, Defense: 2, IsStarter: true, PortraitLabel: "TD", FrameStyle: "bronze"},
		{ID: 6, Name: "ヒーリングバード", Description: "支援寄りの飛行使い。", Rarity: 1, Element: "water", MaxHP: 21, Attack: 7, Defense: 3, IsStarter: true, PortraitLabel: "HB", FrameStyle: "bronze"},
		{ID: 7, Name: "ルーンナイト", Description: "勝利報酬で加入する上位前衛。", Rarity: 3, Element: "earth", MaxHP: 31, Attack: 12, Defense: 5, PortraitLabel: "RK", FrameStyle: "gold"},
		{ID: 8, Name: "クリスタルウィッチ", Description: "高火力の術者。", Rarity: 3, Element: "water", MaxHP: 22, Attack: 14, Defense: 3, PortraitLabel: "CW", FrameStyle: "gold"},
		{ID: 9, Name: "ガーディアンビースト", Description: "重装の守護獣。", Rarity: 3, Element: "light", MaxHP: 34, Attack: 10, Defense: 6, PortraitLabel: "GB", FrameStyle: "gold"},
		{ID: 10, Name: "ブレイズドラグーン", Description: "炎の槍で貫く高火力カード。", Rarity: 4, Element: "fire", MaxHP: 26, Attack: 17, Defense: 4, PortraitLabel: "BD", FrameStyle: "gold"},
		{ID: 11, Name: "アビスアサシン", Description: "闇の一撃に特化した奇襲役。", Rarity: 4, Element: "dark", MaxHP: 23, Attack: 18, Defense: 3, PortraitLabel: "AA", FrameStyle: "gold"},
		{ID: 12, Name: "テンペストフェアリー", Description: "風属性の連続支援役。", Rarity: 2, Element: "wind", MaxHP: 22, Attack: 11, Defense: 4, PortraitLabel: "TF", FrameStyle: "bronze"},
	})
	repo.SeedBosses([]model.Boss{{ID: 1, Name: "遺跡の番人ゴーレム", Description: "古代遺跡の奥で眠っていた石の守護者。地属性で重い一撃と高い耐久を持つ。", Element: "earth", MaxHP: 120, Attack: 13, Defense: 5, RewardExp: 45, RewardCoins: 160, PortraitLabel: "GOLEM", FrameStyle: "boss"}})
	id1, id2, id3, id4, id5 := int64(1), int64(2), int64(3), int64(4), int64(5)
	repo.SeedDrops(1, []repository.WeightedDrop{repository.NewDrop(&id1, "item", 45, 5, 10, "きらめく石を見つけた！"), repository.NewDrop(&id2, "item", 30, 4, 12, "古びたコインを拾った！"), repository.NewDrop(nil, "nothing", 20, 3, 5, "何も見つからなかったが経験を積んだ。"), repository.NewDrop(&id3, "rare_chest", 5, 8, 30, "レア宝箱から幻の羽根を手に入れた！")})
	repo.SeedDrops(2, []repository.WeightedDrop{repository.NewDrop(&id4, "item", 50, 8, 20, "遺跡の欠片を発見した！"), repository.NewDrop(&id2, "item", 20, 6, 15, "古びたコインを見つけた！"), repository.NewDrop(nil, "enemy", 20, 10, 25, "敵をかわしながら奥へ進んだ！"), repository.NewDrop(&id5, "rare_chest", 10, 15, 40, "神秘の宝玉を獲得した！")})
	repo.SeedDrops(3, []repository.WeightedDrop{repository.NewDrop(&id3, "item", 35, 12, 24, "幻の羽根を見つけた！"), repository.NewDrop(&id4, "item", 30, 10, 18, "遺跡の欠片を回収した！"), repository.NewDrop(nil, "enemy", 25, 14, 32, "森の獣を振り切って進んだ！"), repository.NewDrop(&id5, "rare_chest", 10, 20, 60, "神秘の宝玉が眠る宝箱を発見した！")})
	repo.SeedCheckpoints([]model.Checkpoint{
		{
			ID:                "cp-kamisuwa-sta-01",
			QRText:            "QR1",
			Name:              "上諏訪駅",
			Area:              "諏訪エリア",
			Description:       "上諏訪駅チェックポイント。初回報酬と日次報酬を獲得できます。",
			Lat:               36.047050878037396, 
			Lng:               138.11735511690694,
			FirstRewardCoin:   120,
			FirstRewardExp:    20,
			DailyRewardCoin:   30,
			DailyRewardExp:    5,
			EventRewardName:   "上諏訪駅到達ボーナス",
			EventRewardType:   "coin",
			EventRewardValue:  50,
			BossTicketReward:  0,
			GachaTicketReward: 0,
			IsEventActive:     true,
			IsActive:          true,
		},
		{
			ID:                "cp-chino-sta-02",
			QRText:            "QR2",
			Name:              "茅野駅",
			Area:              "茅野エリア",
			Description:       "茅野駅チェックポイント。初回報酬にガチャチケットが含まれます。",
			Lat:               35.99492925295173,
			Lng:               138.15254537977847,
			FirstRewardCoin:   80,
			FirstRewardExp:    25,
			DailyRewardCoin:   25,
			DailyRewardExp:    8,
			EventRewardName:   "茅野駅イベントボーナス",
			EventRewardType:   "gacha_ticket",
			EventRewardValue:  1,
			BossTicketReward:  0,
			GachaTicketReward: 1,
			IsEventActive:     true,
			IsActive:          true,
		},
		{
			ID:                "cp-suwa-tus-03",
			QRText:            "QR3",
			Name:              "公立諏訪東京理科大学",
			Area:              "大学エリア",
			Description:       "公立諏訪東京理科大学チェックポイント。初回報酬にボス挑戦権を獲得できます。",
			Lat:               36.0083526788256,
			Lng:               138.1833705339603,
			FirstRewardCoin:   60,
			FirstRewardExp:    30,
			DailyRewardCoin:   20,
			DailyRewardExp:    10,
			EventRewardName:   "大学到達ボーナス",
			EventRewardType:   "boss_ticket",
			EventRewardValue:  1,
			BossTicketReward:  1,
			GachaTicketReward: 0,
			IsEventActive:     true,
			IsActive:          true,
		},
	})
	now := time.Now()
	repo.SeedNotices([]model.Notice{{ID: 1, Title: "サービス開始のお知らせ", Body: "レトロ探索ゲームの試作サーバが公開されました。", IsPinned: true, PublishedAt: now.Add(-3 * time.Hour), IsActive: true}, {ID: 2, Title: "AUTO BATTLE 実装", Body: "6枚デッキが自動でボスと戦う専用画面を追加しました。", IsPinned: false, PublishedAt: now.Add(-2 * time.Hour), IsActive: true}, {ID: 3, Title: "属性・強化・ガチャ追加", Body: "属性相性、カード強化、デッキ編成、キャラガチャを追加しました。", IsPinned: false, PublishedAt: now.Add(-1 * time.Hour), IsActive: true}, {ID: 4, Title: "チェックポイント報酬追加", Body: "QR1〜QR3 の入力で初回・日次・イベント・挑戦権・ガチャ券の報酬を獲得できるチェックポイント機能を追加しました。", IsPinned: false, PublishedAt: now, IsActive: true}})
}
