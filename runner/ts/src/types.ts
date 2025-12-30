// Created by Yanjunhui

export interface TestCase {
  name: string;
  category: string;
  operation: string;
  collection: string;
  description: string;
  setup?: SetupStep[];
  action: TestAction;
  expected: Expected;
}

export interface SetupStep {
  operation: string;
  data: any;
}

export interface TestAction {
  method: string;
  filter?: any;
  update?: any;
  doc?: any;
  docs?: any[];
  options?: any;
}

export interface Expected {
  count?: number;
  documents?: any[];
  matched_count?: number;
  modified_count?: number;
  deleted_count?: number;
  upserted_id?: any;
  error?: string;
  index_name?: string;
}

export interface TestResult {
  test_name: string;
  language: string;
  mode: string;
  success: boolean;
  error?: string;
  duration_ms: number;
  documents?: any[];
  count?: number;
  matched_count?: number;
  modified_count?: number;
  deleted_count?: number;
  upserted_id?: any;
}

export interface TestSuite {
  version: string;
  generated: string;
  tests: TestCase[];
}

export interface ResultsFile {
  language: string;
  mode: string;
  results: TestResult[];
  summary: Summary;
}

export interface Summary {
  total: number;
  passed: number;
  failed: number;
  skipped: number;
}
