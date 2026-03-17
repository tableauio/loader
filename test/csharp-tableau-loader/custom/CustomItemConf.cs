using System;

namespace Custom
{
    public class CustomItemConf : Tableau.Messager, Tableau.IMessagerName
    {
        private Protoconf.ItemConf.Types.Item? _specialItemConf;

        public string Name() => "CustomItemConf";

        public override bool Load(string dir, Tableau.Format fmt, in Tableau.Load.MessagerOptions? options = null) => true;

        public override bool ProcessAfterLoadAll(in Tableau.Hub hub)
        {
            var itemConf = hub.GetItemConf();
            if (itemConf is null)
            {
                Console.Error.WriteLine("hub get ItemConf failed!");
                return false;
            }
            var conf = itemConf.Get1(1);
            if (conf is null)
            {
                Console.Error.WriteLine("hub get item 1 failed!");
                return false;
            }
            _specialItemConf = conf;
            Console.WriteLine("custom item conf processed");
            return true;
        }

        public string GetSpecialItemName() => _specialItemConf?.Name ?? "";
    }
}