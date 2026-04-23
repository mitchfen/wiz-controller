using Microsoft.AspNetCore.Components;
using WizController.Services;

namespace wizController.Components.Pages;

public partial class Home : IDisposable
{
    [Inject] private WizService Wiz { get; set; } = default!;
    [Inject] private IConfiguration Config { get; set; } = default!;

    private string[] LightIps = Array.Empty<string>();
    private string[] LightNames = Array.Empty<string>();
    private Dictionary<string, int> _brightness = new();
    private int _masterBrightness = 0;
    private CancellationTokenSource? _cts;

    protected override async Task OnInitializedAsync()
    {
        LightIps = Config.GetSection("WizLights:Ips").Get<string[]>() ?? Array.Empty<string>();
        LightNames = Config.GetSection("WizLights:Names").Get<string[]>() ?? Array.Empty<string>();

        if (LightNames.Length != LightIps.Length)
        {
            LightNames = Enumerable.Range(1, LightIps.Length)
                .Select(i => $"Light {i}")
                .ToArray();
        }

        await RefreshStatesAsync();
        _ = PollAsync();
    }

    private async Task PollAsync()
    {
        _cts = new CancellationTokenSource();
        using var timer = new PeriodicTimer(TimeSpan.FromSeconds(5));
        try
        {
            while (await timer.WaitForNextTickAsync(_cts.Token))
            {
                await RefreshStatesAsync();
                await InvokeAsync(StateHasChanged);
            }
        }
        catch (OperationCanceledException) { }
    }

    private async Task RefreshStatesAsync()
    {
        await Task.WhenAll(LightIps.Select(async ip =>
        {
            var state = await Wiz.GetLightStateAsync(ip);
            _brightness[ip] = state.IsOn ? state.Brightness : 0;
        }));
    }

    private async Task SetAllBrightness(int brightness)
    {
        foreach (var ip in LightIps)
        {
            _brightness[ip] = brightness;
            if (brightness == 0)
                await Wiz.SetLightStateAsync(ip, false);
            else
                await Wiz.SetBrightnessAsync(ip, brightness);
        }
    }

    private async Task SetAllBrightnessFromMaster(ChangeEventArgs e)
    {
        if (int.TryParse(e.Value?.ToString(), out int brightness))
        {
            _masterBrightness = brightness;
            await SetAllBrightness(brightness);
        }
    }

    private async Task ChangeBrightness(string ip, ChangeEventArgs e)
    {
        if (int.TryParse(e.Value?.ToString(), out int brightness))
        {
            _brightness[ip] = brightness;
            if (brightness == 0)
                await Wiz.SetLightStateAsync(ip, false);
            else
                await Wiz.SetBrightnessAsync(ip, brightness);
        }
    }

    public void Dispose() => _cts?.Cancel();
}
