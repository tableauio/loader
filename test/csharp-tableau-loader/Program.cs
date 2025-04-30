class Program
{
    static void Main(string[] _)
    {
        Tableau.Registry.Init();

        var options = new Tableau.HubOptions
        {
            Filter = name => name != "TaskConf"
        };
        var hub = new Tableau.Hub(options);
        var loadOptions = new Tableau.Load.Options
        {
            IgnoreUnknownFields = true
        };
        if (!hub.Load("../testdata/conf", Tableau.Format.JSON, loadOptions))
        {
            Console.WriteLine("Failed to load configurations");
            return;
        }

        var taskConf = hub.Get<Tableau.TaskConf>();
        if (taskConf is null)
        {
            Console.WriteLine("TaskConf is null");
        }
        else
        {
            Console.WriteLine($"TaskConf: {taskConf.Data()}");
            Console.WriteLine($"TaskConf Load duration: {taskConf.GetStats().Duration.TotalMilliseconds} ms");
        }

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
            var itemConf2 = hub.GetItemConf();
            Console.WriteLine($"hub.Get<Tableau.ItemConf>() returns same instance with hub.GetItemConf(): {ReferenceEquals(itemConf, itemConf2)}");
            var itemInfoMap = itemConf.FindItemInfoMap();
            if (itemInfoMap != null)
            {
                Console.WriteLine("ItemInfoMap Contents:");
                foreach (var itemPair in itemInfoMap)
                {
                    Console.WriteLine($"  - {itemPair.Key}: ");
                    foreach (var element in itemPair.Value)
                    {
                        Console.WriteLine($"    - {element}");
                    }
                }
            }
        }

        LoadBin();
    }

    static void LoadBin()
    {
        Console.WriteLine("LoadBin");
        var heroConf = new Tableau.HeroConf();
        if (heroConf.Load("../testdata/bin", Tableau.Format.Bin))
        {
            Console.WriteLine($"HeroConf: {heroConf.Data()}");
        }
        if (!heroConf.Load("../testdata/notexist", Tableau.Format.Bin))
        {
            Console.WriteLine("HeroConf not exist");
        }
    }
}