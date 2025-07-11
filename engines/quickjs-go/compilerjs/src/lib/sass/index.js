import * as immutable from "immutable"
import "./sass.dart.js";

const _cliPkgLibrary = globalThis._cliPkgExports.pop();
if (globalThis._cliPkgExports.length === 0) delete globalThis._cliPkgExports;
const _cliPkgExports = {};
_cliPkgLibrary.load({ immutable }, _cliPkgExports);

export const compile = _cliPkgExports.compile;
export const compileAsync = _cliPkgExports.compileAsync;
export const compileString = _cliPkgExports.compileString;
export const compileStringAsync = _cliPkgExports.compileStringAsync;
export const initCompiler = _cliPkgExports.initCompiler;
export const initAsyncCompiler = _cliPkgExports.initAsyncCompiler;
export const Compiler = _cliPkgExports.Compiler;
export const AsyncCompiler = _cliPkgExports.AsyncCompiler;
export const Logger = _cliPkgExports.Logger;
export const SassArgumentList = _cliPkgExports.SassArgumentList;
export const SassBoolean = _cliPkgExports.SassBoolean;
export const SassCalculation = _cliPkgExports.SassCalculation;
export const CalculationOperation = _cliPkgExports.CalculationOperation;
export const CalculationInterpolation = _cliPkgExports.CalculationInterpolation;
export const SassColor = _cliPkgExports.SassColor;
export const SassFunction = _cliPkgExports.SassFunction;
export const SassList = _cliPkgExports.SassList;
export const SassMap = _cliPkgExports.SassMap;
export const SassMixin = _cliPkgExports.SassMixin;
export const SassNumber = _cliPkgExports.SassNumber;
export const SassString = _cliPkgExports.SassString;
export const Value = _cliPkgExports.Value;
export const CustomFunction = _cliPkgExports.CustomFunction;
export const ListSeparator = _cliPkgExports.ListSeparator;
export const sassFalse = _cliPkgExports.sassFalse;
export const sassNull = _cliPkgExports.sassNull;
export const sassTrue = _cliPkgExports.sassTrue;
export const Exception = _cliPkgExports.Exception;
export const PromiseOr = _cliPkgExports.PromiseOr;
export const info = _cliPkgExports.info;
export const render = _cliPkgExports.render;
export const renderSync = _cliPkgExports.renderSync;
export const TRUE = _cliPkgExports.TRUE;
export const FALSE = _cliPkgExports.FALSE;
export const NULL = _cliPkgExports.NULL;
export const types = _cliPkgExports.types;
export const NodePackageImporter = _cliPkgExports.NodePackageImporter;
export const deprecations = _cliPkgExports.deprecations;
export const Version = _cliPkgExports.Version;
