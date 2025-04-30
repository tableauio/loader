namespace Tableau
{
    public enum Format
    {
        Unknown,
        JSON,
        Bin
    }

    public static class Util
    {

        private const string _unknownExt = ".unknown";
        private const string _jsonExt = ".json";
        private const string _binExt = ".binpb";

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