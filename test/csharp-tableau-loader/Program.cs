using System;
using Google.Protobuf.WellKnownTypes;
using Protoconf;
using Tableau;

class Program
{
    static void Main(string[] args)
    {
        Tableau.Registry.Init();

        var options = new HubOptions
        {
            Filter = name => true
        };
        var hub = new Tableau.Hub(options);
        var ok = hub.Load("../testdata/conf", Tableau.Format.JSON);
        Console.WriteLine($"Load result: {ok}");

        var heroConf = hub.Get<Tableau.HeroConf>();
        if (heroConf is null)
        {
            Console.WriteLine("HeroConf is null");
        }
        else
        {
            Console.WriteLine($"HeroConf: {heroConf.Data()}");
            Console.WriteLine($"HeroConf Load duration: {heroConf.GetStats().Duration.TotalMilliseconds} ms");
            // Traverse top-level OrderedMap (HeroOrderedMap)
            var heroOrderedMap = heroConf.GetOrderedMap();
            if (heroOrderedMap != null)
            {
                Console.WriteLine("Hero OrderedMap:");
                foreach (var heroPair in heroOrderedMap)
                {
                    Console.WriteLine($"Hero: {heroPair.Key}");
                    Console.WriteLine($"  - Hero Data: {heroPair.Value.Item2}");
                    // Traverse nested Attr OrderedMap
                    var attrOrderedMap = heroPair.Value.Item1;
                    if (attrOrderedMap != null && attrOrderedMap.Count > 0)
                    {
                        Console.WriteLine("  Attributes:");
                        foreach (var attrPair in attrOrderedMap)
                        {
                            Console.WriteLine($"    - {attrPair.Key}: {attrPair.Value}");
                        }
                    }
                }
            }
        }

        var itemConf = hub.Get<Tableau.ItemConf>();
        if (itemConf is null)
        {
            Console.WriteLine("ItemConf is null");
        }
        else
        {
            Console.WriteLine($"ItemConf: {itemConf.Data()}");
            Console.WriteLine($"ItemConf Load duration: {itemConf.GetStats().Duration.TotalMilliseconds} ms");
        }

        itemConf = hub.GetItemConf();
        if (itemConf is null)
        {
            Console.WriteLine("ItemConf is null");
        }
        else
        {
            Console.WriteLine($"ItemConf: {itemConf.Data()}");
            Console.WriteLine($"ItemConf Load duration: {itemConf.GetStats().Duration.TotalMilliseconds} ms");
        }
    }
}