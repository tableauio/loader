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
    }
}