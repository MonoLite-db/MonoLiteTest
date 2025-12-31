// Created by Yanjunhui

/**
 * 测试用例接口，定义单个测试的完整结构
 * // EN: Test case interface, defines the complete structure of a single test
 */
export interface TestCase {
  /** 测试名称 // EN: Test name */
  name: string;
  /** 测试类别 // EN: Test category */
  category: string;
  /** 操作类型 // EN: Operation type */
  operation: string;
  /** 集合名称 // EN: Collection name */
  collection: string;
  /** 测试描述 // EN: Test description */
  description: string;
  /** 前置步骤（可选） // EN: Setup steps (optional) */
  setup?: SetupStep[];
  /** 测试动作 // EN: Test action */
  action: TestAction;
  /** 期望结果 // EN: Expected result */
  expected: Expected;
}

/**
 * 前置步骤接口，定义测试前的准备操作
 * // EN: Setup step interface, defines preparation operations before test
 */
export interface SetupStep {
  /** 操作类型 // EN: Operation type */
  operation: string;
  /** 操作数据 // EN: Operation data */
  data: any;
}

/**
 * 测试动作接口，定义要执行的数据库操作
 * // EN: Test action interface, defines the database operation to execute
 */
export interface TestAction {
  /** 方法名称 // EN: Method name */
  method: string;
  /** 查询过滤条件（可选） // EN: Query filter (optional) */
  filter?: any;
  /** 更新内容（可选） // EN: Update content (optional) */
  update?: any;
  /** 单个文档（可选） // EN: Single document (optional) */
  doc?: any;
  /** 多个文档（可选） // EN: Multiple documents (optional) */
  docs?: any[];
  /** 操作选项（可选） // EN: Operation options (optional) */
  options?: any;
}

/**
 * 期望结果接口，定义测试的预期输出
 * // EN: Expected result interface, defines the expected output of the test
 */
export interface Expected {
  /** 文档数量 // EN: Document count */
  count?: number;
  /** 返回的文档列表 // EN: Returned document list */
  documents?: any[];
  /** 匹配的文档数量 // EN: Matched document count */
  matched_count?: number;
  /** 修改的文档数量 // EN: Modified document count */
  modified_count?: number;
  /** 删除的文档数量 // EN: Deleted document count */
  deleted_count?: number;
  /** 插入或更新的文档ID // EN: Upserted document ID */
  upserted_id?: any;
  /** 期望的错误信息 // EN: Expected error message */
  error?: string;
  /** 索引名称 // EN: Index name */
  index_name?: string;
}

/**
 * 测试结果接口，记录单个测试的执行结果
 * // EN: Test result interface, records the execution result of a single test
 */
export interface TestResult {
  /** 测试名称 // EN: Test name */
  test_name: string;
  /** 编程语言 // EN: Programming language */
  language: string;
  /** 运行模式 // EN: Running mode */
  mode: string;
  /** 是否成功 // EN: Whether succeeded */
  success: boolean;
  /** 错误信息（可选） // EN: Error message (optional) */
  error?: string;
  /** 执行耗时（毫秒） // EN: Execution duration (milliseconds) */
  duration_ms: number;
  /** 返回的文档列表 // EN: Returned document list */
  documents?: any[];
  /** 文档数量 // EN: Document count */
  count?: number;
  /** 匹配的文档数量 // EN: Matched document count */
  matched_count?: number;
  /** 修改的文档数量 // EN: Modified document count */
  modified_count?: number;
  /** 删除的文档数量 // EN: Deleted document count */
  deleted_count?: number;
  /** 插入或更新的文档ID // EN: Upserted document ID */
  upserted_id?: any;
}

/**
 * 测试套件接口，包含所有测试用例
 * // EN: Test suite interface, contains all test cases
 */
export interface TestSuite {
  /** 版本号 // EN: Version number */
  version: string;
  /** 生成时间 // EN: Generation time */
  generated: string;
  /** 测试用例列表 // EN: Test case list */
  tests: TestCase[];
}

/**
 * 结果文件接口，定义测试结果的输出格式
 * // EN: Results file interface, defines the output format of test results
 */
export interface ResultsFile {
  /** 编程语言 // EN: Programming language */
  language: string;
  /** 运行模式 // EN: Running mode */
  mode: string;
  /** 测试结果列表 // EN: Test result list */
  results: TestResult[];
  /** 测试摘要 // EN: Test summary */
  summary: Summary;
}

/**
 * 测试摘要接口，汇总测试执行情况
 * // EN: Test summary interface, summarizes test execution status
 */
export interface Summary {
  /** 总测试数 // EN: Total test count */
  total: number;
  /** 通过数 // EN: Passed count */
  passed: number;
  /** 失败数 // EN: Failed count */
  failed: number;
  /** 跳过数 // EN: Skipped count */
  skipped: number;
}
