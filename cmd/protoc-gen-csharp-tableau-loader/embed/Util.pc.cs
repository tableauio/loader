using System;
using System.IO;
namespace Tableau
{
    /// <summary>
    /// Format specifies the format of the configuration file.
    /// </summary>
    public enum Format
    {
        Unknown,
        JSON,
        Bin
    }

    /// <summary>
    /// Util provides common utility functions.
    /// </summary>
    public static class Util
    {
        [ThreadStatic] private static string? _errMsg;

        /// <summary>
        /// Get the last error message.
        /// </summary>
        public static string GetErrMsg() => _errMsg ?? "";

        /// <summary>
        /// Set the error message.
        /// </summary>
        public static void SetErrMsg(string msg) => _errMsg = msg;

        private const string _unknownExt = ".unknown";
        private const string _jsonExt = ".json";
        private const string _binExt = ".binpb";

        /// <summary>
        /// GetFormat returns the Format type determined by the file extension of the given path.
        /// </summary>
        public static Format GetFormat(string path)
        {
            string ext = Path.GetExtension(path);
            return ext switch
            {
                _jsonExt => Format.JSON,
                _binExt => Format.Bin,
                _ => Format.Unknown,
            };
        }

        /// <summary>
        /// Format2Ext returns the file extension corresponding to the given format.
        /// </summary>
        public static string Format2Ext(Format fmt)
        {
            return fmt switch
            {
                Format.JSON => _jsonExt,
                Format.Bin => _binExt,
                _ => _unknownExt,
            };
        }
    }
}