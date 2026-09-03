-- ============================================================
-- 健康服务模块：评估量表 mock 数据（health_assessment_forms）
-- 生成 20 条常用评估量表，status=1（启用），version=1
-- 结构：name 量表名 / dimension 评估维度 / description 说明
--       questions [{key,title,options:[{label,score}]}]
--       score_rule [{min,max,level,conclusion}]
-- 幂等：按 name 不存在才插入
-- ============================================================

INSERT INTO health_assessment_forms (name, dimension, description, questions, score_rule, version, status)
SELECT t.name, t.dimension, t.description, t.questions, t.score_rule, t.version, t.status
FROM (
  SELECT '日常生活活动能力评估(ADL)' AS name, 'adl' AS dimension, '评估老年人及功能障碍者进食、洗漱、穿衣、如厕、行走等日常生活自理能力。' AS description,
         '[{"key":"q_adl_1","title":"进食能力","options":[{"label":"完全独立","score":25},{"label":"需他人照看","score":15},{"label":"需喂食","score":5}]},{"key":"q_adl_2","title":"穿衣洗漱","options":[{"label":"完全独立","score":20},{"label":"需协助","score":10},{"label":"完全依赖","score":0}]},{"key":"q_adl_3","title":"如厕控制","options":[{"label":"完全独立","score":25},{"label":"偶尔失禁","score":15},{"label":"经常失禁","score":5}]},{"key":"q_adl_4","title":"行走活动","options":[{"label":"行动自如","score":30},{"label":"需助行器","score":18},{"label":"卧床不起","score":6}]}]' AS questions,
         '[{"min":0,"max":20,"level":"重度依赖","conclusion":"日常生活完全依赖他人照护，建议安排专人陪护。"},{"min":21,"max":60,"level":"中度依赖","conclusion":"部分生活自理需要协助，宜引入日常照护与康复训练。"},{"min":61,"max":100,"level":"基本自理","conclusion":"生活基本自理，保持现有生活方式并定期评估。"}]' AS score_rule, 1 AS version, 1 AS status
UNION ALL SELECT '认知功能评估(MMSE简化)', 'cognition', '筛查认知障碍，评估记忆、定向、注意与计算等能力。',
  '[{"key":"q_cog_1","title":"定向力：今天星期几","options":[{"label":"回答正确","score":20},{"label":"回答错误","score":0}]},{"key":"q_cog_2","title":"即时记忆：记住三个词","options":[{"label":"全部记住","score":20},{"label":"部分记住","score":10},{"label":"无法记住","score":0}]},{"key":"q_cog_3","title":"注意力与计算：100-7","options":[{"label":"连续计算正确","score":30},{"label":"有错误","score":15},{"label":"无法完成","score":0}]},{"key":"q_cog_4","title":"回忆：说出刚才三个词","options":[{"label":"全部说出","score":30},{"label":"部分说出","score":15},{"label":"说不出","score":0}]}]' ,
  '[{"min":0,"max":35,"level":"重度障碍","conclusion":"认知功能明显受损，建议线下神经/精神科进一步评估。"},{"min":36,"max":70,"level":"中度障碍","conclusion":"存在认知下降，建议进行专业认知功能量表复核。"},{"min":71,"max":100,"level":"正常","conclusion":"认知功能基本正常，保持规律用脑与健康生活方式。"}]', 1, 1
UNION ALL SELECT '老年抑郁自评量表(GDS-简化)', 'mood', '评估老年人近两周情绪低落、兴趣减退等抑郁倾向。',
  '[{"key":"q_dep_1","title":"您是否经常感到情绪低落","options":[{"label":"从不","score":0},{"label":"偶尔","score":10},{"label":"经常","score":30}]},{"key":"q_dep_2","title":"您是否对以前喜欢的事失去兴趣","options":[{"label":"没有","score":0},{"label":"有时","score":15},{"label":"明显","score":35}]},{"key":"q_dep_3","title":"您是否觉得生活没有意义","options":[{"label":"不会","score":0},{"label":"偶尔这样想","score":10},{"label":"经常这样想","score":35}]}]' ,
  '[{"min":0,"max":20,"level":"正常","conclusion":"情绪状态良好，建议保持社交与运动。"},{"min":21,"max":50,"level":"轻度抑郁","conclusion":"存在轻度抑郁倾向，建议增加陪伴并心理疏导。"},{"min":51,"max":100,"level":"中重度抑郁","conclusion":"抑郁症状明显，建议尽快转介专业心理机构。"}]', 1, 1
UNION ALL SELECT '广泛性焦虑自评量表(GAD简化)', 'mood', '评估近两周过度担忧、紧张、坐立不安等焦虑症状。',
  '[{"key":"q_anx_1","title":"是否会无缘无故感到紧张","options":[{"label":"不会","score":0},{"label":"有时","score":15},{"label":"经常","score":35}]},{"key":"q_anx_2","title":"是否难以放松或入睡困难","options":[{"label":"没有","score":0},{"label":"有时","score":15},{"label":"经常","score":35}]},{"key":"q_anx_3","title":"是否总是担心不好的事发生","options":[{"label":"不会","score":0},{"label":"偶尔","score":10},{"label":"经常","score":30}]}]' ,
  '[{"min":0,"max":25,"level":"正常","conclusion":"未见明显焦虑症状。"},{"min":26,"max":55,"level":"轻度焦虑","conclusion":"存在轻度焦虑，建议放松训练与规律作息。"},{"min":56,"max":100,"level":"中重度焦虑","conclusion":"焦虑症状明显，建议转介专业心理咨询。"}]', 1, 1
UNION ALL SELECT '跌倒风险评估(Morse简化)', 'safety', '评估跌倒风险，指导居家防跌倒改造与陪护安排。',
  '[{"key":"q_fall_1","title":"近期是否有跌倒史","options":[{"label":"无","score":0},{"label":"近3个月1次","score":20},{"label":"多次","score":40}]},{"key":"q_fall_2","title":"是否存在步态不稳","options":[{"label":"步态稳健","score":0},{"label":"稍有摇晃","score":20},{"label":"明显不稳/需助行","score":40}]},{"key":"q_fall_3","title":"是否服用影响平衡的药物","options":[{"label":"无","score":0},{"label":"有","score":20}]}]' ,
  '[{"min":0,"max":30,"level":"低风险","conclusion":"跌倒风险较低，注意居家防滑。"},{"min":31,"max":65,"level":"中风险","conclusion":"存在中等跌倒风险，建议地面防滑、加装扶手。"},{"min":66,"max":100,"level":"高风险","conclusion":"跌倒风险高，建议贴身陪护并整体居家适老化改造。"}]', 1, 1
UNION ALL SELECT '吞咽功能风险评估', 'dysphagia', '评估进食呛咳、吞咽困难，防范误吸风险。',
  '[{"key":"q_sw_1","title":"饮水或进食是否呛咳","options":[{"label":"从不","score":0},{"label":"偶尔","score":30},{"label":"经常","score":50}]},{"key":"q_sw_2","title":"吞咽后是否常有食物残留","options":[{"label":"无","score":0},{"label":"偶尔","score":25},{"label":"常有","score":50}]}]' ,
  '[{"min":0,"max":30,"level":"正常","conclusion":"吞咽功能正常，正常饮食即可。"},{"min":31,"max":60,"level":"轻度风险","conclusion":"建议细嚼慢咽、食物切小块，进食时保持坐位。"},{"min":61,"max":100,"level":"高风险","conclusion":"误吸风险高，建议稠化流质并转介吞咽康复。"}]', 1, 1
UNION ALL SELECT '疼痛数字评分(NRS-11)', 'pain', '以0-10数字量化疼痛强度，评估部位与影响。',
  '[{"key":"q_pain_1","title":"当前疼痛程度为","options":[{"label":"无痛(0)","score":0},{"label":"轻度(1-3)","score":30},{"label":"中度(4-6)","score":60},{"label":"重度(7-10)","score":100}]}]' ,
  '[{"min":0,"max":10,"level":"无痛","conclusion":"未见明显疼痛。"},{"min":11,"max":40,"level":"轻度疼痛","conclusion":"轻微疼痛，注意休息与局部护理。"},{"min":41,"max":70,"level":"中度疼痛","conclusion":"中度疼痛影响生活，建议就医明确病因。"},{"min":71,"max":100,"level":"重度疼痛","conclusion":"重度疼痛，建议尽快安排就医或疼痛门诊。"}]', 1, 1
UNION ALL SELECT '营养风险筛查(NRS2002简化)', 'nutrition', '筛查营养不良风险，评估体重下降与进食情况。',
  '[{"key":"q_nut_1","title":"近3个月体重是否下降","options":[{"label":"无下降","score":0},{"label":"下降5%以内","score":20},{"label":"下降超过5%","score":40}]},{"key":"q_nut_2","title":"近期进食量变化","options":[{"label":"正常","score":0},{"label":"减少1/3","score":20},{"label":"减少一半及以上","score":40}]},{"key":"q_nut_3","title":"是否患有影响进食的疾病","options":[{"label":"无","score":0},{"label":"有","score":20}]}]' ,
  '[{"min":0,"max":25,"level":"营养正常","conclusion":"营养状态良好，保持均衡饮食。"},{"min":26,"max":60,"level":"营养风险","conclusion":"存在营养不良风险，建议加强优质蛋白摄入。"},{"min":61,"max":100,"level":"高度营养风险","conclusion":"营养风险高，建议营养师评估并制定膳食方案。"}]', 1, 1
UNION ALL SELECT '压疮风险评分(Braden简化)', 'skin', '评估皮肤受压风险，用于卧床老人预防压疮。',
  '[{"key":"q_brad_1","title":"感知能力与皮肤反应","options":[{"label":"无受损","score":0},{"label":"部分受限","score":20},{"label":"严重受限","score":40}]},{"key":"q_brad_2","title":"活动能力与卧床时间","options":[{"label":"可自主活动","score":0},{"label":"偶尔离床","score":20},{"label":"长期卧床","score":40}]},{"key":"q_brad_3","title":"皮肤潮湿程度","options":[{"label":"保持干燥","score":0},{"label":"偶尔潮湿","score":10},{"label":"持续潮湿","score":20}]}]' ,
  '[{"min":0,"max":25,"level":"低风险","conclusion":"压疮风险低，注意变换体位即可。"},{"min":26,"max":60,"level":"中风险","conclusion":"定期翻身边减压，使用减压垫。"},{"min":61,"max":100,"level":"高风险","conclusion":"压疮风险高，需勤翻身、专业皮肤护理与减压辅具。"}]', 1, 1
UNION ALL SELECT '睡眠质量评估(PSQI简化)', 'sleep', '评估入睡、夜间觉醒、晨起状态等整体睡眠质量。',
  '[{"key":"q_slp_1","title":"入睡所需时间","options":[{"label":"小于30分钟","score":0},{"label":"30-60分钟","score":20},{"label":"超过60分钟","score":40}]},{"key":"q_slp_2","title":"夜间觉醒次数","options":[{"label":"基本不醒","score":0},{"label":"1-2次","score":20},{"label":"3次及以上","score":40}]},{"key":"q_slp_3","title":"晨起精神状况","options":[{"label":"精神充沛","score":0},{"label":"一般","score":10},{"label":"疲惫睏倦","score":20}]}]' ,
  '[{"min":0,"max":25,"level":"睡眠良好","conclusion":"睡眠质量良好，保持规律作息。"},{"min":26,"max":60,"level":"中度障碍","conclusion":"睡眠质量下降，建议睡前放松、减少咖啡因。"},{"min":61,"max":100,"level":"重度障碍","conclusion":"存在明显睡眠障碍，建议就医或助眠干预。"}]', 1, 1
UNION ALL SELECT '肌力与握力评估', 'motor', '评估四肢肌力与整体活动耐力，反映肌肉衰减风险。',
  '[{"key":"q_ms_1","title":"能否独立从椅子上站起","options":[{"label":"轻松完成","score":0},{"label":"需扶手","score":30},{"label":"无法完成","score":60}]},{"key":"q_ms_2","title":"上手臂抬举能力","options":[{"label":"轻松抬举","score":0},{"label":"略感费力","score":20},{"label":"无法抬起","score":40}]}]' ,
  '[{"min":0,"max":30,"level":"肌力正常","conclusion":"肌力基本正常，建议坚持力量锻炼。"},{"min":31,"max":70,"level":"轻度衰减","conclusion":"存在肌力下降，建议增加抗阻训练与蛋白质摄入。"},{"min":71,"max":100,"level":"明显衰减","conclusion":"肌力衰减明显，建议康复评估并系统训练。"}]', 1, 1
UNION ALL SELECT '平衡能力评估(Berg简化)', 'balance', '评估静态站立与转身平衡，识别跌倒高危人群。',
  '[{"key":"q_bal_1","title":"无支撑站立30秒","options":[{"label":"轻松站立","score":0},{"label":"略有晃动","score":30},{"label":"无法站立","score":60}]},{"key":"q_bal_2","title":"转身360度能力","options":[{"label":"平稳转身","score":0},{"label":"需扶持","score":20},{"label":"无法完成","score":40}]}]' ,
  '[{"min":0,"max":30,"level":"平衡良好","conclusion":"平衡能力良好，保持锻炼。"},{"min":31,"max":70,"level":"轻度失衡","conclusion":"平衡略有下降，宜进行平衡训练并注意防滑。"},{"min":71,"max":100,"level":"明显失衡","conclusion":"平衡能力差，跌倒概率高，建议康复训练并行适老化改造。"}]', 1, 1
UNION ALL SELECT '社会支持与孤独感评估(UCLA简化)', 'social', '评估社交参与与孤独感，用于心理关爱与社群关怀。',
  '[{"key":"q_soc_1","title":"您与他人交往的频率","options":[{"label":"频繁","score":0},{"label":"偶尔","score":30},{"label":"极少","score":60}]},{"key":"q_soc_2","title":"是否常感到孤独无助","options":[{"label":"从不","score":0},{"label":"有时","score":20},{"label":"经常","score":40}]}]' ,
  '[{"min":0,"max":30,"level":"社会支持良好","conclusion":"社交状况良好，继续保持人际往来。"},{"min":31,"max":70,"level":"轻度孤独","conclusion":"存在轻度孤独感，建议增加社区活动与陪伴。"},{"min":71,"max":100,"level":"明显孤独","conclusion":"孤独感明显，建议纳入定期探访与心理关怀服务。"}]', 1, 1
UNION ALL SELECT '慢性病自我管理能力评估', 'disease', '评估慢病患者在服药、监测、饮食等方面的自我管理能力。',
  '[{"key":"q_cdm_1","title":"能否按时按量服药","options":[{"label":"总是","score":0},{"label":"经常","score":15},{"label":"较少","score":40}]},{"key":"q_cdm_2","title":"能否定期监测血压/血糖","options":[{"label":"规律监测","score":0},{"label":"偶尔","score":20},{"label":"从不","score":40}]},{"key":"q_cdm_3","title":"是否遵从医嘱控制饮食","options":[{"label":"严格执行","score":0},{"label":"基本执行","score":15},{"label":"难以执行","score":25}]}]' ,
  '[{"min":0,"max":25,"level":"管理良好","conclusion":"自我管理良好，巩固现有习惯。"},{"min":26,"max":60,"level":"基本达标","conclusion":"部分管理欠佳，建议加强服药与监测提醒。"},{"min":61,"max":100,"level":"管理不足","conclusion":"自我管理不足，建议纳入健康管理随访与用药辅导。"}]', 1, 1
UNION ALL SELECT '用药依从性评估(Morisky简化)', 'medication', '评估遵医嘱服药程度，发现漏服、停药风险。',
  '[{"key":"q_med_1","title":"是否有时忘记服药","options":[{"label":"从不","score":0},{"label":"偶尔","score":30},{"label":"经常","score":60}]},{"key":"q_med_2","title":"感觉好转后是否自行停药","options":[{"label":"从不","score":0},{"label":"偶尔","score":40}]}]' ,
  '[{"min":0,"max":35,"level":"依从性良好","conclusion":"服药依从性良好。"},{"min":36,"max":100,"level":"依从性不足","conclusion":"存在漏服或自行停药风险，建议用药管理提醒与家属协助。"}]', 1, 1
UNION ALL SELECT '尿失禁风险与泌尿功能评估', 'excretory', '评估尿失禁程度与对生活的影响，指导护理方案。',
  '[{"key":"q_uri_1","title":"尿液不自主漏出频率","options":[{"label":"从不","score":0},{"label":"偶尔","score":30},{"label":"经常","score":60}]},{"key":"q_uri_2","title":"漏尿对日常生活的影响","options":[{"label":"无影响","score":0},{"label":"有一定影响","score":20},{"label":"严重影响","score":40}]}]' ,
  '[{"min":0,"max":30,"level":"功能正常","conclusion":"泌尿功能正常。"},{"min":31,"max":70,"level":"轻度失禁","conclusion":"存在轻度失禁，建议盆底训练与定时如厕。"},{"min":71,"max":100,"level":"明显失禁","conclusion":"失禁明显，建议使用护理用品并评估膀胱管理。"}]', 1, 1
UNION ALL SELECT '口腔健康状态评估', 'oral', '评估口腔清洁、牙列与咀嚼功能，防范口腔疾病。',
  '[{"key":"q_mou_1","title":"牙齿数量与咀嚼能力","options":[{"label":"全口/多数健全","score":0},{"label":"部分缺损","score":30},{"label":"多数缺失","score":60}]},{"key":"q_mou_2","title":"口腔清洁习惯及是否有不适","options":[{"label":"清洁良好无不适","score":0},{"label":"清洁一般偶有疼痛","score":20},{"label":"清洁差常疼痛","score":40}]}]' ,
  '[{"min":0,"max":30,"level":"口腔健康","conclusion":"口腔状况良好，保持清洁。"},{"min":31,"max":70,"level":"中度问题","conclusion":"存在口腔问题，建议加强清洁并口腔科检查。"},{"min":71,"max":100,"level":"明显问题","conclusion":"口腔问题明显影响进食，建议尽快口腔诊疗。"}]', 1, 1
UNION ALL SELECT '视听力功能评估', 'sensory', '评估视听功能对日常交流与安全的影响。',
  '[{"key":"q_sen_1","title":"视力是否影响日常活动","options":[{"label":"不影响","score":0},{"label":"轻度影响","score":30},{"label":"明显影响","score":60}]},{"key":"q_sen_2","title":"听力是否影响与人交流","options":[{"label":"不影响","score":0},{"label":"需大声重复","score":20},{"label":"明显困难","score":40}]}]' ,
  '[{"min":0,"max":30,"level":"功能良好","conclusion":"视听力基本良好。"},{"min":31,"max":70,"level":"轻度受损","conclusion":"视听力轻度受损，建议配置老花镜/助听器评估。"},{"min":71,"max":100,"level":"明显受损","conclusion":"视听力明显受损，影响安全和交流，建议专科评估。"}]', 1, 1
UNION ALL SELECT '居家环境安全评估', 'environment', '评估居家地面、照明、卫浴、通道安全，识别隐患。',
  '[{"key":"q_env_1","title":"地面是否平坦防滑","options":[{"label":"平坦防滑","score":0},{"label":"局部湿滑","score":30},{"label":"明显湿滑不平","score":60}]},{"key":"q_env_2","title":"卫浴是否配备扶手防滑设施","options":[{"label":"齐全","score":0},{"label":"部分具备","score":20},{"label":"没有","score":40}]}]' ,
  '[{"min":0,"max":30,"level":"环境安全","conclusion":"居家环境较安全，保持现状。"},{"min":31,"max":70,"level":"一般隐患","conclusion":"存在安全隐患，建议局部整改（防滑垫、扶手、照明）。"},{"min":71,"max":100,"level":"明显隐患","conclusion":"安全隐患明显，建议进行全面适老化改造。"}]', 1, 1
UNION ALL SELECT '自我照护与工具性日常活动评估(IADL)', 'iadl', '评估使用电话、购物、理财、出行等工具性日常活动能力。',
  '[{"key":"q_iadl_1","title":"能否独立使用电话、购物等","options":[{"label":"完全独立","score":0},{"label":"需协助","score":30},{"label":"无法独立","score":60}]},{"key":"q_iadl_2","title":"能否独立管理服药与财务","options":[{"label":"完全独立","score":0},{"label":"需提醒协助","score":20},{"label":"无法独立","score":40}]}]' ,
  '[{"min":0,"max":30,"level":"独立能力强","conclusion":"工具性日常活动基本独立。"},{"min":31,"max":70,"level":"部分依赖","conclusion":"部分需协助，建议提供代办或家属辅助。"},{"min":71,"max":100,"level":"明显依赖","conclusion":"明显依赖他人，建议安排生活照料与陪伴服务。"}]', 1, 1
) t
WHERE NOT EXISTS (SELECT 1 FROM health_assessment_forms f WHERE f.name = t.name);

-- 校验：统计当前启用量表数量
SELECT COUNT(*) AS total_enabled_forms FROM health_assessment_forms WHERE status = 1;