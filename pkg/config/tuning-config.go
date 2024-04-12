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
	ShuffleDataset *bool    `json:"shuffle_dataset,omitempty"`
	ShuffleSeed    *int     `json:"shuffle_seed,omitempty"`
	ContextColumn  *string  `json:"context_column,omitempty"`
	ResponseColumn *string  `json:"response_column,omitempty"`
	TrainTestSplit *float64 `json:"train_test_split,omitempty"`
}

type TokenizerParams struct {
	AddSpecialTokens        *bool   `json:"add_special_tokens,omitempty"`
	Padding                 *bool   `json:"padding,omitempty"`
	Truncation              *bool   `json:"truncation,omitempty"`
	MaxLength               *int    `json:"max_length,omitempty"`
	Stride                  *int    `json:"stride,omitempty"`
	IsSplitIntoWords        *bool   `json:"is_split_into_words,omitempty"`
	PadToMultipleOf         *int    `json:"pad_to_multiple_of,omitempty"`
	ReturnTensors           *string `json:"return_tensors,omitempty"`
	ReturnTokenTypeIds      *bool   `json:"return_token_type_ids,omitempty"`
	ReturnAttentionMask     *bool   `json:"return_attention_mask,omitempty"`
	ReturnOverflowingTokens *bool   `json:"return_overflowing_tokens,omitempty"`
	ReturnSpecialTokensMask *bool   `json:"return_special_tokens_mask,omitempty"`
	ReturnOffsetsMapping    *bool   `json:"return_offsets_mapping,omitempty"`
	ReturnLength            *bool   `json:"return_length,omitempty"`
	Verbose                 *bool   `json:"verbose,omitempty"`
}

type ModelConfig struct {
	PretrainedModelNameOrPath *string `json:"pretrained_model_name_or_path,omitempty"`
	// TODO: Consider adding support for complex types like maps if necessary.
	//StateDict                 *map[string]interface{} `json:"state_dict,omitempty"`
	CacheDir          *string `json:"cache_dir,omitempty"`
	FromTF            *bool   `json:"from_tf,omitempty"`
	ForceDownload     *bool   `json:"force_download,omitempty"`
	ResumeDownload    *bool   `json:"resume_download,omitempty"`
	Proxies           *string `json:"proxies,omitempty"`
	OutputLoadingInfo *bool   `json:"output_loading_info,omitempty"`
	LocalFilesOnly    *bool   `json:"local_files_only,omitempty"`
	Revision          *string `json:"revision,omitempty"`
	TrustRemoteCode   *bool   `json:"trust_remote_code,omitempty"`
	LoadIn4bit        *bool   `json:"load_in_4bit,omitempty"`
	LoadIn8bit        *bool   `json:"load_in_8bit,omitempty"`
	TorchDtype        *string `json:"torch_dtype,omitempty"`
	DeviceMap         *string `json:"device_map,omitempty"`
}

type QuantizationConfig struct {
	QuantMethod                 *string   `json:"quant_method,omitempty"`
	LoadIn8bit                  *bool     `json:"load_in_8bit,omitempty"`
	LoadIn4bit                  *bool     `json:"load_in_4bit,omitempty"`
	LLMInt8Threshold            *float64  `json:"llm_int8_threshold,omitempty"`
	LLMInt8SkipModules          *[]string `json:"llm_int8_skip_modules,omitempty"`
	LLMInt8EnableFP32CPUOffload *bool     `json:"llm_int8_enable_fp32_cpu_offload,omitempty"`
	LLMInt8HasFP16Weight        *bool     `json:"llm_int8_has_fp16_weight,omitempty"`
	BNB4bitComputeDtype         *string   `json:"bnb_4bit_compute_dtype,omitempty"`
	BNB4bitQuantType            *string   `json:"bnb_4bit_quant_type,omitempty"`
	BNB4bitUseDoubleQuant       *bool     `json:"bnb_4bit_use_double_quant,omitempty"`
}

// LoraConfig represents the extended LoRa configuration with additional fields for customization.
type LoraConfig struct {
	R                 *int      `json:"r,omitempty"`
	LoraAlpha         *int      `json:"lora_alpha,omitempty"`
	LoraDropout       *float64  `json:"lora_dropout,omitempty"`
	FanInFanOut       *bool     `json:"fan_in_fan_out,omitempty"`
	Bias              *string   `json:"bias,omitempty"`
	UseRSLora         *bool     `json:"use_rslora,omitempty"`
	ModulesToSave     *[]string `json:"modules_to_save,omitempty"`
	InitLoraWeights   *bool     `json:"init_lora_weights,omitempty"`
	TargetModules     *[]string `json:"target_modules,omitempty"`
	LayersToTransform *[]int    `json:"layers_to_transform,omitempty"`
	LayersPattern     *[]string `json:"layers_pattern,omitempty"`
	// TODO: Consider adding support for complex types like maps if necessary.
	// RankPattern       map[string]int `json:"rank_pattern,omitempty"`
	// AlphaPattern      map[string]int `json:"alpha_pattern,omitempty"`
}

// TrainingArguments represents the training arguments for a model.
type TrainingArguments struct {
	OutputDir                 string   `json:"output_dir"`
	OverwriteOutputDir        *bool    `json:"overwrite_output_dir"`
	DoTrain                   *bool    `json:"do_train"`
	DoEval                    *bool    `json:"do_eval"`
	DoPredict                 *bool    `json:"do_predict"`
	EvaluationStrategy        *string  `json:"evaluation_strategy"`
	PredictionLossOnly        *bool    `json:"prediction_loss_only"`
	PerDeviceTrainBatchSize   *int     `json:"per_device_train_batch_size"`
	PerDeviceEvalBatchSize    *int     `json:"per_device_eval_batch_size"`
	GradientAccumulationSteps *int     `json:"gradient_accumulation_steps"`
	EvalAccumulationSteps     *int     `json:"eval_accumulation_steps,omitempty"`
	EvalDelay                 *float64 `json:"eval_delay,omitempty"`
	LearningRate              *float64 `json:"learning_rate"`
	WeightDecay               *float64 `json:"weight_decay"`
	AdamBeta1                 *float64 `json:"adam_beta1"`
	AdamBeta2                 *float64 `json:"adam_beta2"`
	AdamEpsilon               *float64 `json:"adam_epsilon"`
	MaxGradNorm               *float64 `json:"max_grad_norm"`
	NumTrainEpochs            *float64 `json:"num_train_epochs"`
	MaxSteps                  *int     `json:"max_steps"`
	LrSchedulerType           *string  `json:"lr_scheduler_type"`
	//LrSchedulerKwargs           *map[string]interface{} `json:"lr_scheduler_kwargs,omitempty"`
	WarmupRatio                 *float64 `json:"warmup_ratio"`
	WarmupSteps                 *int     `json:"warmup_steps"`
	LogLevel                    *string  `json:"log_level"`
	LogLevelReplica             *string  `json:"log_level_replica"`
	LogOnEachNode               *bool    `json:"log_on_each_node"`
	LoggingDir                  *string  `json:"logging_dir,omitempty"`
	LoggingStrategy             *string  `json:"logging_strategy"`
	LoggingFirstStep            *bool    `json:"logging_first_step"`
	LoggingSteps                *float64 `json:"logging_steps"`
	LoggingNanInfFilter         *bool    `json:"logging_nan_inf_filter"`
	SaveStrategy                *string  `json:"save_strategy"`
	SaveSteps                   *float64 `json:"save_steps"`
	SaveTotalLimit              *int     `json:"save_total_limit,omitempty"`
	SaveSafeTensors             *bool    `json:"save_safetensors,omitempty"`
	SaveOnEachNode              *bool    `json:"save_on_each_node"`
	SaveOnlyModel               *bool    `json:"save_only_model"`
	UseCPU                      *bool    `json:"use_cpu"`
	Seed                        *int     `json:"seed"`
	DataSeed                    *int     `json:"data_seed,omitempty"`
	JitModeEval                 *bool    `json:"jit_mode_eval"`
	UseIPEX                     *bool    `json:"use_ipex"`
	Bf16                        *bool    `json:"bf16"`
	Fp16                        *bool    `json:"fp16"`
	Fp16OptLevel                *string  `json:"fp16_opt_level"`
	Fp16Backend                 *string  `json:"fp16_backend,omitempty"`
	HalfPrecisionBackend        *string  `json:"half_precision_backend"`
	Bf16FullEval                *bool    `json:"bf16_full_eval"`
	Fp16FullEval                *bool    `json:"fp16_full_eval"`
	Tf32                        *bool    `json:"tf32,omitempty"`
	LocalRank                   *int     `json:"local_rank"`
	DdpBackend                  *string  `json:"ddp_backend,omitempty"`
	TpuNumCores                 *int     `json:"tpu_num_cores,omitempty"`
	DataloaderDropLast          *bool    `json:"dataloader_drop_last"`
	EvalSteps                   *float64 `json:"eval_steps,omitempty"`
	DataloaderNumWorkers        *int     `json:"dataloader_num_workers"`
	PastIndex                   *int     `json:"past_index,omitempty"`
	RunName                     *string  `json:"run_name,omitempty"`
	DisableTqdm                 *bool    `json:"disable_tqdm"`
	RemoveUnusedColumns         *bool    `json:"remove_unused_columns"`
	LabelNames                  []string `json:"label_names,omitempty"`
	LoadBestModelAtEnd          *bool    `json:"load_best_model_at_end"`
	MetricForBestModel          *string  `json:"metric_for_best_model,omitempty"`
	GreaterIsBetter             *bool    `json:"greater_is_better"`
	IgnoreDataSkip              *bool    `json:"ignore_data_skip"`
	Fsdp                        *string  `json:"fsdp,omitempty"`        // or use []string for list of options
	FsdpConfig                  *string  `json:"fsdp_config,omitempty"` // or map[string]interface{} if it's complex
	Deepspeed                   *string  `json:"deepspeed,omitempty"`
	AcceleratorConfig           *string  `json:"accelerator_config,omitempty"`
	LabelSmoothingFactor        *float64 `json:"label_smoothing_factor"`
	Debug                       *string  `json:"debug,omitempty"`
	Optim                       *string  `json:"optim,omitempty"`
	OptimArgs                   *string  `json:"optim_args,omitempty"`
	GroupByLength               *bool    `json:"group_by_length"`
	LengthColumnName            *string  `json:"length_column_name,omitempty"`
	ReportTo                    *string  `json:"report_to,omitempty"`
	DdpFindUnusedParameters     *bool    `json:"ddp_find_unused_parameters"`
	DdpBucketCapMb              *int     `json:"ddp_bucket_cap_mb,omitempty"`
	DdpBroadcastBuffers         *bool    `json:"ddp_broadcast_buffers"`
	DataloaderPinMemory         *bool    `json:"dataloader_pin_memory"`
	DataloaderPersistentWorkers *bool    `json:"dataloader_persistent_workers"`
	DataloaderPrefetchFactor    *int     `json:"dataloader_prefetch_factor,omitempty"`
	SkipMemoryMetrics           *bool    `json:"skip_memory_metrics"`
	PushToHub                   *bool    `json:"push_to_hub"`
	ResumeFromCheckpoint        *string  `json:"resume_from_checkpoint,omitempty"`
	HubModelID                  *string  `json:"hub_model_id,omitempty"`
	HubStrategy                 *string  `json:"hub_strategy"`
	HubToken                    *string  `json:"hub_token,omitempty"`
	HubPrivateRepo              *bool    `json:"hub_private_repo"`
	HubAlwaysPush               *bool    `json:"hub_always_push"`
	GradientCheckpointing       *bool    `json:"gradient_checkpointing"`
	//TODO: GradientCheckpointingKwargs *map[string]interface{} `json:"gradient_checkpointing_kwargs,omitempty"`
	IncludeInputsForMetrics   *bool    `json:"include_inputs_for_metrics"`
	AutoFindBatchSize         *bool    `json:"auto_find_batch_size"`
	FullDeterminism           *bool    `json:"full_determinism"`
	TorchDynamo               *string  `json:"torchdynamo,omitempty"`
	RayScope                  *string  `json:"ray_scope"`
	DdpTimeout                *int     `json:"ddp_timeout"`
	UseMPSDevice              *bool    `json:"use_mps_device"`
	TorchCompile              *bool    `json:"torch_compile"`
	TorchCompileBackend       *string  `json:"torch_compile_backend,omitempty"`
	TorchCompileMode          *string  `json:"torch_compile_mode,omitempty"`
	SplitBatches              *bool    `json:"split_batches,omitempty"`
	IncludeTokensPerSecond    *bool    `json:"include_tokens_per_second,omitempty"`
	IncludeNumInputTokensSeen *bool    `json:"include_num_input_tokens_seen,omitempty"`
	NeftuneNoiseAlpha         *float64 `json:"neftune_noise_alpha,omitempty"`
	// TODO: Can provide acceptable string values for structs
	// EvaluationStrategy, LoggingStrategy, SaveStrategy, RayScope
}

type DataCollator struct {
	//Tokenizer             PreTrainedTokenizer `json:"tokenizer"`
	Mlm             *bool    `json:"mlm,omitempty"`
	MlmProbability  *float64 `json:"mlm_probability,omitempty"`
	PadToMultipleOf *int     `json:"pad_to_multiple_of,omitempty"`
	ReturnTensors   *string  `json:"return_tensors,omitempty"`
}
