package config

 type TrainingConfig struct {
 	ModelConfig        *ModelConfig        `yaml:"ModelConfig"`
 	TokenizerParams    *TokenizerParams    `yaml:"TokenizerParams"`
 	QuantizationConfig *QuantizationConfig `yaml:"QuantizationConfig"`
 	LoraConfig         *LoraConfig         `yaml:"LoraConfig"`
 	TrainingArguments  *TrainingArguments  `yaml:"TrainingArguments"`
 	DatasetConfig      *DatasetConfig      `yaml:"DatasetConfig"`
 	DataCollator       *DataCollator       `yaml:"DataCollator"`
 }

 type DatasetConfig struct {
 	ShuffleDataset *bool    `yaml:"shuffle_dataset,omitempty"`
 	ShuffleSeed    *int     `yaml:"shuffle_seed,omitempty"`
 	ContextColumn  *string  `yaml:"context_column,omitempty"`
 	ResponseColumn *string  `yaml:"response_column,omitempty"`
 	TrainTestSplit *float64 `yaml:"train_test_split,omitempty"`
 }

 type TokenizerParams struct {
 	AddSpecialTokens        *bool   `yaml:"add_special_tokens,omitempty"`
 	Padding                 *bool   `yaml:"padding,omitempty"`
 	Truncation              *bool   `yaml:"truncation,omitempty"`
 	MaxLength               *int    `yaml:"max_length,omitempty"`
 	Stride                  *int    `yaml:"stride,omitempty"`
 	IsSplitIntoWords        *bool   `yaml:"is_split_into_words,omitempty"`
 	PadToMultipleOf         *int    `yaml:"pad_to_multiple_of,omitempty"`
 	ReturnTensors           *string `yaml:"return_tensors,omitempty"`
 	ReturnTokenTypeIds      *bool   `yaml:"return_token_type_ids,omitempty"`
 	ReturnAttentionMask     *bool   `yaml:"return_attention_mask,omitempty"`
 	ReturnOverflowingTokens *bool   `yaml:"return_overflowing_tokens,omitempty"`
 	ReturnSpecialTokensMask *bool   `yaml:"return_special_tokens_mask,omitempty"`
 	ReturnOffsetsMapping    *bool   `yaml:"return_offsets_mapping,omitempty"`
 	ReturnLength            *bool   `yaml:"return_length,omitempty"`
 	Verbose                 *bool   `yaml:"verbose,omitempty"`
 }

 type ModelConfig struct {
 	PretrainedModelNameOrPath *string `yaml:"pretrained_model_name_or_path,omitempty"`
 	// TODO: Consider adding support for complex types like maps if necessary.
 	//StateDict                 *map[string]interface{} `yaml:"state_dict,omitempty"`
 	CacheDir          *string `yaml:"cache_dir,omitempty"`
 	FromTF            *bool   `yaml:"from_tf,omitempty"`
 	ForceDownload     *bool   `yaml:"force_download,omitempty"`
 	ResumeDownload    *bool   `yaml:"resume_download,omitempty"`
 	Proxies           *string `yaml:"proxies,omitempty"`
 	OutputLoadingInfo *bool   `yaml:"output_loading_info,omitempty"`
 	LocalFilesOnly    *bool   `yaml:"local_files_only,omitempty"`
 	Revision          *string `yaml:"revision,omitempty"`
 	TrustRemoteCode   *bool   `yaml:"trust_remote_code,omitempty"`
 	LoadIn4bit        *bool   `yaml:"load_in_4bit,omitempty"`
 	LoadIn8bit        *bool   `yaml:"load_in_8bit,omitempty"`
 	TorchDtype        *string `yaml:"torch_dtype,omitempty"`
 	DeviceMap         *string `yaml:"device_map,omitempty"`
 }

 type QuantizationConfig struct {
 	QuantMethod                 *string   `yaml:"quant_method,omitempty"`
 	LoadIn8bit                  *bool     `yaml:"load_in_8bit,omitempty"`
 	LoadIn4bit                  *bool     `yaml:"load_in_4bit,omitempty"`
 	LLMInt8Threshold            *float64  `yaml:"llm_int8_threshold,omitempty"`
 	LLMInt8SkipModules          *[]string `yaml:"llm_int8_skip_modules,omitempty"`
 	LLMInt8EnableFP32CPUOffload *bool     `yaml:"llm_int8_enable_fp32_cpu_offload,omitempty"`
 	LLMInt8HasFP16Weight        *bool     `yaml:"llm_int8_has_fp16_weight,omitempty"`
 	BNB4bitComputeDtype         *string   `yaml:"bnb_4bit_compute_dtype,omitempty"`
 	BNB4bitQuantType            *string   `yaml:"bnb_4bit_quant_type,omitempty"`
 	BNB4bitUseDoubleQuant       *bool     `yaml:"bnb_4bit_use_double_quant,omitempty"`
 }

 // LoraConfig represents the extended LoRa configuration with additional fields for customization.
 type LoraConfig struct {
 	R                 *int      `yaml:"r,omitempty"`
 	LoraAlpha         *int      `yaml:"lora_alpha,omitempty"`
 	LoraDropout       *float64  `yaml:"lora_dropout,omitempty"`
 	FanInFanOut       *bool     `yaml:"fan_in_fan_out,omitempty"`
 	Bias              *string   `yaml:"bias,omitempty"`
 	UseRSLora         *bool     `yaml:"use_rslora,omitempty"`
 	ModulesToSave     *[]string `yaml:"modules_to_save,omitempty"`
 	InitLoraWeights   *bool     `yaml:"init_lora_weights,omitempty"`
 	TargetModules     *[]string `yaml:"target_modules,omitempty"`
 	LayersToTransform *[]int    `yaml:"layers_to_transform,omitempty"`
 	LayersPattern     *[]string `yaml:"layers_pattern,omitempty"`
 	// TODO: Consider adding support for complex types like maps if necessary.
 	// RankPattern       map[string]int `yaml:"rank_pattern,omitempty"`
 	// AlphaPattern      map[string]int `yaml:"alpha_pattern,omitempty"`
 }

 // TrainingArguments represents the training arguments for a model.
 type TrainingArguments struct {
 	OutputDir                 string   `yaml:"output_dir"`
 	OverwriteOutputDir        *bool    `yaml:"overwrite_output_dir"`
 	DoTrain                   *bool    `yaml:"do_train"`
 	DoEval                    *bool    `yaml:"do_eval"`
 	DoPredict                 *bool    `yaml:"do_predict"`
 	EvaluationStrategy        *string  `yaml:"evaluation_strategy"`
 	PredictionLossOnly        *bool    `yaml:"prediction_loss_only"`
 	PerDeviceTrainBatchSize   *int     `yaml:"per_device_train_batch_size"`
 	PerDeviceEvalBatchSize    *int     `yaml:"per_device_eval_batch_size"`
 	GradientAccumulationSteps *int     `yaml:"gradient_accumulation_steps"`
 	EvalAccumulationSteps     *int     `yaml:"eval_accumulation_steps,omitempty"`
 	EvalDelay                 *float64 `yaml:"eval_delay,omitempty"`
 	LearningRate              *float64 `yaml:"learning_rate"`
 	WeightDecay               *float64 `yaml:"weight_decay"`
 	AdamBeta1                 *float64 `yaml:"adam_beta1"`
 	AdamBeta2                 *float64 `yaml:"adam_beta2"`
 	AdamEpsilon               *float64 `yaml:"adam_epsilon"`
 	MaxGradNorm               *float64 `yaml:"max_grad_norm"`
 	NumTrainEpochs            *float64 `yaml:"num_train_epochs"`
 	MaxSteps                  *int     `yaml:"max_steps"`
 	LrSchedulerType           *string  `yaml:"lr_scheduler_type"`
 	//LrSchedulerKwargs           *map[string]interface{} `yaml:"lr_scheduler_kwargs,omitempty"`
 	WarmupRatio                 *float64 `yaml:"warmup_ratio"`
 	WarmupSteps                 *int     `yaml:"warmup_steps"`
 	LogLevel                    *string  `yaml:"log_level"`
 	LogLevelReplica             *string  `yaml:"log_level_replica"`
 	LogOnEachNode               *bool    `yaml:"log_on_each_node"`
 	LoggingDir                  *string  `yaml:"logging_dir,omitempty"`
 	LoggingStrategy             *string  `yaml:"logging_strategy"`
 	LoggingFirstStep            *bool    `yaml:"logging_first_step"`
 	LoggingSteps                *float64 `yaml:"logging_steps"`
 	LoggingNanInfFilter         *bool    `yaml:"logging_nan_inf_filter"`
 	SaveStrategy                *string  `yaml:"save_strategy"`
 	SaveSteps                   *float64 `yaml:"save_steps"`
 	SaveTotalLimit              *int     `yaml:"save_total_limit,omitempty"`
 	SaveSafeTensors             *bool    `yaml:"save_safetensors,omitempty"`
 	SaveOnEachNode              *bool    `yaml:"save_on_each_node"`
 	SaveOnlyModel               *bool    `yaml:"save_only_model"`
 	UseCPU                      *bool    `yaml:"use_cpu"`
 	Seed                        *int     `yaml:"seed"`
 	DataSeed                    *int     `yaml:"data_seed,omitempty"`
 	JitModeEval                 *bool    `yaml:"jit_mode_eval"`
 	UseIPEX                     *bool    `yaml:"use_ipex"`
 	Bf16                        *bool    `yaml:"bf16"`
 	Fp16                        *bool    `yaml:"fp16"`
 	Fp16OptLevel                *string  `yaml:"fp16_opt_level"`
 	Fp16Backend                 *string  `yaml:"fp16_backend,omitempty"`
 	HalfPrecisionBackend        *string  `yaml:"half_precision_backend"`
 	Bf16FullEval                *bool    `yaml:"bf16_full_eval"`
 	Fp16FullEval                *bool    `yaml:"fp16_full_eval"`
 	Tf32                        *bool    `yaml:"tf32,omitempty"`
 	LocalRank                   *int     `yaml:"local_rank"`
 	DdpBackend                  *string  `yaml:"ddp_backend,omitempty"`
 	TpuNumCores                 *int     `yaml:"tpu_num_cores,omitempty"`
 	DataloaderDropLast          *bool    `yaml:"dataloader_drop_last"`
 	EvalSteps                   *float64 `yaml:"eval_steps,omitempty"`
 	DataloaderNumWorkers        *int     `yaml:"dataloader_num_workers"`
 	PastIndex                   *int     `yaml:"past_index,omitempty"`
 	RunName                     *string  `yaml:"run_name,omitempty"`
 	DisableTqdm                 *bool    `yaml:"disable_tqdm"`
 	RemoveUnusedColumns         *bool    `yaml:"remove_unused_columns"`
 	LabelNames                  []string `yaml:"label_names,omitempty"`
 	LoadBestModelAtEnd          *bool    `yaml:"load_best_model_at_end"`
 	MetricForBestModel          *string  `yaml:"metric_for_best_model,omitempty"`
 	GreaterIsBetter             *bool    `yaml:"greater_is_better"`
 	IgnoreDataSkip              *bool    `yaml:"ignore_data_skip"`
 	Fsdp                        *string  `yaml:"fsdp,omitempty"`        // or use []string for list of options
 	FsdpConfig                  *string  `yaml:"fsdp_config,omitempty"` // or map[string]interface{} if it's complex
 	Deepspeed                   *string  `yaml:"deepspeed,omitempty"`
 	AcceleratorConfig           *string  `yaml:"accelerator_config,omitempty"`
 	LabelSmoothingFactor        *float64 `yaml:"label_smoothing_factor"`
 	Debug                       *string  `yaml:"debug,omitempty"`
 	Optim                       *string  `yaml:"optim,omitempty"`
 	OptimArgs                   *string  `yaml:"optim_args,omitempty"`
 	GroupByLength               *bool    `yaml:"group_by_length"`
 	LengthColumnName            *string  `yaml:"length_column_name,omitempty"`
 	ReportTo                    *string  `yaml:"report_to,omitempty"`
 	DdpFindUnusedParameters     *bool    `yaml:"ddp_find_unused_parameters"`
 	DdpBucketCapMb              *int     `yaml:"ddp_bucket_cap_mb,omitempty"`
 	DdpBroadcastBuffers         *bool    `yaml:"ddp_broadcast_buffers"`
 	DataloaderPinMemory         *bool    `yaml:"dataloader_pin_memory"`
 	DataloaderPersistentWorkers *bool    `yaml:"dataloader_persistent_workers"`
 	DataloaderPrefetchFactor    *int     `yaml:"dataloader_prefetch_factor,omitempty"`
 	SkipMemoryMetrics           *bool    `yaml:"skip_memory_metrics"`
 	PushToHub                   *bool    `yaml:"push_to_hub"`
 	ResumeFromCheckpoint        *string  `yaml:"resume_from_checkpoint,omitempty"`
 	HubModelID                  *string  `yaml:"hub_model_id,omitempty"`
 	HubStrategy                 *string  `yaml:"hub_strategy"`
 	HubToken                    *string  `yaml:"hub_token,omitempty"`
 	HubPrivateRepo              *bool    `yaml:"hub_private_repo"`
 	HubAlwaysPush               *bool    `yaml:"hub_always_push"`
 	GradientCheckpointing       *bool    `yaml:"gradient_checkpointing"`
 	//TODO: GradientCheckpointingKwargs *map[string]interface{} `yaml:"gradient_checkpointing_kwargs,omitempty"`
 	IncludeInputsForMetrics   *bool    `yaml:"include_inputs_for_metrics"`
 	AutoFindBatchSize         *bool    `yaml:"auto_find_batch_size"`
 	FullDeterminism           *bool    `yaml:"full_determinism"`
 	TorchDynamo               *string  `yaml:"torchdynamo,omitempty"`
 	RayScope                  *string  `yaml:"ray_scope"`
 	DdpTimeout                *int     `yaml:"ddp_timeout"`
 	UseMPSDevice              *bool    `yaml:"use_mps_device"`
 	TorchCompile              *bool    `yaml:"torch_compile"`
 	TorchCompileBackend       *string  `yaml:"torch_compile_backend,omitempty"`
 	TorchCompileMode          *string  `yaml:"torch_compile_mode,omitempty"`
 	SplitBatches              *bool    `yaml:"split_batches,omitempty"`
 	IncludeTokensPerSecond    *bool    `yaml:"include_tokens_per_second,omitempty"`
 	IncludeNumInputTokensSeen *bool    `yaml:"include_num_input_tokens_seen,omitempty"`
 	NeftuneNoiseAlpha         *float64 `yaml:"neftune_noise_alpha,omitempty"`
 	// TODO: Can provide acceptable string values for structs
 	// EvaluationStrategy, LoggingStrategy, SaveStrategy, RayScope
 }

 type DataCollator struct {
 	//Tokenizer             PreTrainedTokenizer `yaml:"tokenizer"`
 	Mlm             *bool    `yaml:"mlm,omitempty"`
 	MlmProbability  *float64 `yaml:"mlm_probability,omitempty"`
 	PadToMultipleOf *int     `yaml:"pad_to_multiple_of,omitempty"`
 	ReturnTensors   *string  `yaml:"return_tensors,omitempty"`
 }